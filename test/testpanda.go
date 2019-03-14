
package apps

import (
	"20.26.28.57/panda/client/prometheus"
	"20.26.28.57/panda/model"
	"context"
	"fmt"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promModel "github.com/prometheus/common/model"
	"strconv"
	"sync"
	"time"
)

//const FederalPromeLabel  = "aaa"
type AppProfInfo struct {
	AppBase
	DataCenter   []string
	clusterInfo  []ClusterConfig
	ClusterLabel string `json:"cluster_label,omitempty"`
	//ClusterName  string
}
type RespProfData struct {
	Data AppProfilingInfo `json:"data"`
	//ErrCount string             `json:"err_count"`
}

type AppProfilingInfo struct {
	ClusterLabel string `json:"cluster_label"`
	//ClusterName  string        `json:"clustername"`
	PeInfo
	RateInfo
	TotalInfo
}

type PeInfo struct {
	P50  float64 `json:"p50"`
	P90  float64 `json:"p90"`
	P99  float64 `json:"p99"`
	P995 float64 `json:"p99.5"`
}
type RateInfo struct {
	//SuccRate float64
	FailRate float64 `json:"fail_rate"`
}
type TotalInfo struct {
	FailRequestsTotal float64 `json:"fail_requests_total"`
	RequestsTotal     float64 `json:"requests_total"`
}

//fr appstatus

func (a *AppProfInfo) GetAppProfiling() (*AppProfilingInfo, error) {

	var err error
	namespace, appCluster, _, _, err := model.GetAppInfo(a.AppId)
	if err != nil {
		return nil, err
	}
	if namespace == "" && appCluster == "" {
		err = fmt.Errorf("AppId: %s 不存在", a.AppId)
		return nil, err
	}
	a.clusterInfo, _ = a.toJsonCluster(appCluster)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg = &sync.WaitGroup{} //控制子线程的任务执行,等同于Map/Reduce处理
	wg.Add(3)
	api, err := prometheus.GetFederalPromAPI()
	if err != nil {
		logger.Errorf("prometheus.GetPromAPI err: %+v", err)
	}
	resp := AppProfilingInfo{}
	if a.ClusterLabel != "" {
		for _, val := range a.clusterInfo {
			label := val.Label
			if a.ClusterLabel != label {
				continue
			} else {
				resp.ClusterLabel = label
				go getPeMetrics(ctx, api, wg, &resp, a.AppId, a.ClusterLabel, namespace)
				go getRateMetrics(ctx, api, wg, &resp, a.AppId, a.ClusterLabel, namespace)
				go getTotalMetrics(ctx, api, wg, &resp, a.AppId, a.ClusterLabel, namespace)
			}
		}
	} else {
		resp.ClusterLabel = ""
		go getPeMetrics(ctx, api, wg, &resp, a.AppId, a.ClusterLabel, namespace)
		go getRateMetrics(ctx, api, wg, &resp, a.AppId, a.ClusterLabel, namespace)
		go getTotalMetrics(ctx, api, wg, &resp, a.AppId, a.ClusterLabel, namespace)
	}

	wg.Wait()
	logger.Printf("finish clusdata %+v \n", resp)
	return &resp, nil

}

var FailTotalPromQuery = "sum (istio_requests_total{response_code!~'2.*',destination_service='%s'})"
var TotalPromQuery = "sum (istio_requests_total{destination_service='%s'})"
var ClusterFailTotalPromQuery = "sum (istio_requests_total{response_code!~'2.*',destination_service='%s',cluster='%s'})"
var ClusterTotalPromQuery = "sum (istio_requests_total{destination_service='%s',cluster='%s'})"
var FailRatePromQuery = "sum (rate (istio_requests_total{response_code!~'2.*',destination_service='%s'}[1m]))"
var SuccRatePromQuery = "sum (rate (istio_requests_total{response_code=~'2.*',destination_service='%s'}[1m]))"
var ClusterFailRatePromQuery = "sum (rate (istio_requests_total{response_code!~'2.*',destination_service='%s',cluster='%s'}[1m]))"
var ClusterSuccRatePromQuery = "sum (rate (istio_requests_total{response_code=~'2.*',destination_service='%s',cluster='%s'}[1m]))"
var PePromQuery = "histogram_quantile (%s, sum (irate (istio_request_duration_seconds_bucket{destination_service='%s'}[1m])) by (destination_service,le))"
var ClusterPePromQuery = "histogram_quantile (%s, sum (irate (istio_request_duration_seconds_bucket{destination_service='%s',cluster='%s'}[1m])) by (destination_service,le))"

//获取相应数据的handler

func getPeMetrics(ctx context.Context, api promv1.API, wg *sync.WaitGroup, redata *AppProfilingInfo, appid, cluster, namespace string) (err error) {
	svc := fmt.Sprintf("%s-svc.%s.svc.cluster.local", appid, namespace)
	defer wg.Done()
	if cluster != "" {
		redata.P50, err = promDealData(ctx, api, appid, generateClusterPeQuery("0.50", ClusterPePromQuery, svc, cluster))
		redata.P90, err = promDealData(ctx, api, appid, generateClusterPeQuery("0.90", ClusterPePromQuery, svc, cluster))
		redata.P99, err = promDealData(ctx, api, appid, generateClusterPeQuery("0.99", ClusterPePromQuery, svc, cluster))
		redata.P995, err = promDealData(ctx, api, appid, generateClusterPeQuery("0.995", ClusterPePromQuery, svc, cluster))
	} else {
		redata.P50, err = promDealData(ctx, api, appid, generatePeQuery("0.50", PePromQuery, svc))
		redata.P90, err = promDealData(ctx, api, appid, generatePeQuery("0.90", PePromQuery, svc))
		redata.P99, err = promDealData(ctx, api, appid, generatePeQuery("0.99", PePromQuery, svc))
		redata.P995, err = promDealData(ctx, api, appid, generatePeQuery("0.995", PePromQuery, svc))
	}

	return

}

func getRateMetrics(ctx context.Context, api promv1.API, wg *sync.WaitGroup, redata *AppProfilingInfo, appid, cluster, namespace string) (err error) {
	var fpara string
	svc := fmt.Sprintf("%s-svc.%s.svc.cluster.local", appid, namespace)
	defer wg.Done()
	if cluster != "" {
		//spara = fmt.Sprintf(ClusterSuccRatePromQuery, svc, cluster)
		fpara = fmt.Sprintf(ClusterFailRatePromQuery, svc, cluster)
	} else {
		//spara = fmt.Sprintf(FailRatePromQuery, svc)
		fpara = fmt.Sprintf(FailRatePromQuery, svc)
	}
	redata.FailRate, err = promDealData(ctx, api, appid, fpara)
	//redata.SuccRate, err = promDealData(ctx, api, appid, spara)
	return
}

func getTotalMetrics(ctx context.Context, api promv1.API, wg *sync.WaitGroup, redata *AppProfilingInfo, appid, cluster, namespace string) (err error) {
	var tpara, fpara string
	svc := fmt.Sprintf("%s-svc.%s.svc.cluster.local", appid, namespace)
	defer wg.Done()
	if cluster != "" {
		tpara = fmt.Sprintf(ClusterTotalPromQuery, svc, cluster)
		fpara = fmt.Sprintf(ClusterFailTotalPromQuery, svc, cluster)
	} else {
		tpara = fmt.Sprintf(TotalPromQuery, svc)
		fpara = fmt.Sprintf(FailTotalPromQuery, svc)
	}

	redata.FailRequestsTotal, err = promDealData(ctx, api, appid, fpara)
	redata.RequestsTotal, err = promDealData(ctx, api, appid, tpara)
	//fmt.Printf("finish FailRequestsTotal %f \n",redata.FailRequestsTotal)
	//fmt.Printf("finish RequestsTotal %f \n",redata.RequestsTotal)
	return
}

//组装PSQL并对返回整理成所需格式

func promDealData(ctx context.Context, api promv1.API, appid, psql string) (tval float64, err error) {

	result, err := promData(ctx, api, psql)
	if err != nil {
		logger.Errorf("appId:%s err: %s", appid, err)
		return
	}
	if vectors, ok := result.(promModel.Vector); ok {
		for _, v := range vectors {
			//logger.Debugf("mem:%f,v val:%f", mem, float64(v.Value))
			tdata := fmt.Sprintf("%0.5f", float64(v.Value))
			if tdata == "NaN" {
				tdata = "0"
			}
			tval, err = strconv.ParseFloat(tdata, 64)
			if err == nil {
				return
			}
		}
	} else {
		err = fmt.Errorf("GetPeData prometheus query result is not vector, appid %s", appid)
		logger.Error(err)
	}
	return
}

func promData(ctx context.Context, api promv1.API, basepsql string) (promModel.Value, error) {

	result, err := api.Query(ctx, basepsql, time.Now())
	if err != nil {
		err = fmt.Errorf("apiquery basql::%s err:%s", basepsql, err)
		return nil, err
	}
	return result, nil

}

func generatePeQuery(qua, basql, appId string) string {
	return fmt.Sprintf(basql, qua, appId)
}
func generateClusterPeQuery(qua, basql, appId, cluster string) string {
	return fmt.Sprintf(basql, qua, appId, cluster)
}

//func generateDestSvc(appId string) string {
//	return fmt.Sprintf("%s.mtest.svc.cluster.local",appId)
//}

