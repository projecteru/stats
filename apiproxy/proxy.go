package apiproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/projecteru2/stats/config"
	"github.com/projecteru2/stats/handler"
)

// PodsMemCap proxy to http://citadel.test.ricebook.net/api/v1/pod/<podname>/getmemcap
func PodsMemCap() (map[string]map[string]interface{}, error) {
	pods, err := handler.CorePods()
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}
	wg.Add(len(pods))

	type info struct {
		PodName string
		Info    map[string]interface{}
	}
	podMemCapInfoChan := make(chan info)
	podMemCapInfos := make(map[string]map[string]interface{}, 0)

	go func() {
		wg.Add(1)
		defer wg.Done()
		remaining := len(pods)
		for i := range podMemCapInfoChan {
			podMemCapInfos[i.PodName] = i.Info
			if remaining--; remaining == 0 {
				close(podMemCapInfoChan)
			}
		}

	}()

	for _, pod := range pods {
		go func(podname string) {
			defer wg.Done()
			uri := fmt.Sprintf("pod/%s/getmemcap", podname)
			r, err := doReq(uri)
			if err != nil {
				log.Errorf("get pod memcap error: %s", err)
			}
			i := info{
				PodName: podname,
				Info:    r,
			}
			podMemCapInfoChan <- i
		}(pod)
	}

	wg.Wait()

	return podMemCapInfos, nil
}

func doReq(uri string) (map[string]interface{}, error) {
	citadelHost := config.C.Citadel.Host
	citadelAuthToken := config.C.Citadel.Auth

	url := fmt.Sprintf("%s/api/v1/%s", citadelHost, uri)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Neptulon-Token", citadelAuthToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	v := make(map[string]interface{})
	json.Unmarshal(buf.Bytes(), &v)

	return v, nil
}
