package adfunc

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	admissionv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	register(AdmissionFunc{
		Type: AdmissionTypeMutating,
		Path: "/oversold",
		Func: oversold,
	})
}

func oversold(request *admissionv1.AdmissionRequest) (*admissionv1.AdmissionResponse, error) {
	//获取属性Kind为Node
	switch request.Kind.Kind {
	case "Node":
		node := v1.Node{}
		if err := jsoniter.Unmarshal(request.Object.Raw, &node); err != nil {
			errMsg := fmt.Sprintf("[route.Mutating] /oversold: failed to unmarshal object: %v", err)
			klog.Error(errMsg)
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Code:    http.StatusBadRequest,
					Message: errMsg,
				},
			}, nil
		}
		//判断是否该节点需要进行超售
		if node.Labels["kubernetes.io/oversold"] != "oversold" {

			return &admissionv1.AdmissionResponse{
				Allowed:   true,
				PatchType: JSONPatch(),
				Result: &metav1.Status{
					Code:    http.StatusOK,
					Message: "节点无需超售",
				},
			}, nil
		}

		klog.Info(request.UserInfo.Username + " 的label为kubernetes.io/oversold:" + node.Labels["kubernetes.io/oversold"])
		klog.Info(request.UserInfo.Username + " ===================该节点允许超售========================")
		patches := []Patch{
			{
				Option: PatchOptionReplace,
				Path:   "/status/allocatable/cpu",
				//Value: "32",
				Value: overcpu(Quantitytostring(node.Status.Allocatable.Cpu()), node.Labels["kubernetes.io/overcpu"]),
			},
			{
				Option: PatchOptionReplace,
				Path:   "/status/allocatable/memory",
				//Value:  "134217728Ki",
				Value: overmem(Quantitytostring(node.Status.Allocatable.Memory()), node.Labels["kubernetes.io/overmem"]),
			},
		}
		// 实际可分配资源
		klog.Info(request.UserInfo.Username + " oldCPU: " + Quantitytostring(node.Status.Allocatable.Cpu()))
		klog.Info(request.UserInfo.Username + " oldMem: " + Quantitytostring(node.Status.Allocatable.Memory()))
		// 超卖系数
		klog.Info(request.UserInfo.Username + " CPUFactor: " + node.Labels["kubernetes.io/overcpu"])
		klog.Info(request.UserInfo.Username + " MemFactor: " + node.Labels["kubernetes.io/overmem"])
		// 超卖后可分配资源
		klog.Info(request.UserInfo.Username + " overSoldCPU: " + overcpu(Quantitytostring(node.Status.Allocatable.Cpu()), node.Labels["kubernetes.io/overcpu"]))
		klog.Info(request.UserInfo.Username + " overSoldMem: " + overmem(Quantitytostring(node.Status.Allocatable.Memory()), node.Labels["kubernetes.io/overmem"]))
		patch, err := jsoniter.Marshal(patches)
		if err != nil {
			errMsg := fmt.Sprintf("[route.Mutating] /oversold: failed to marshal patch: %v", err)
			logger.Error(errMsg)
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Code:    http.StatusInternalServerError,
					Message: errMsg,
				},
			}, nil
		}
		logger.Infof("[route.Mutating] /oversold: patches: %s", string(patch))
		return &admissionv1.AdmissionResponse{
			Allowed:   true,
			Patch:     patch,
			PatchType: JSONPatch(),
			Result: &metav1.Status{
				Code:    http.StatusOK,
				Message: "success",
			},
		}, nil

	default:
		errMsg := fmt.Sprintf("[route.Mutating] /oversold: received wrong kind request: %s, Only support Kind: Deployment", request.Kind.Kind)
		logger.Error(errMsg)
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Code:    http.StatusForbidden,
				Message: errMsg,
			},
		}, nil
	}

}

// *resource.Quantity类型转string
func Quantitytostring(r *resource.Quantity) string {
	return fmt.Sprint(r)
}

// cpu 超售计算
func overcpu(cpu, multiple string) string {
	if multiple == "" {
		multiple = "1"
	}
	//b, _ := strconv.Atoi(multiple)
	b, _ := strconv.ParseFloat(multiple, 64)

	// cpu的值可能是 32 或者 29800m
	if strings.HasSuffix(cpu, "m") {
		a, err := strconv.ParseFloat(strings.Trim(cpu, "m"), 64)
		if err != nil {
			klog.Error("--------内存超售计算失败----------")
			klog.Error(err)
			return "1"
		}
		return strconv.Itoa(int(a*b)) + "m"
	} else {
		a, _ := strconv.ParseFloat(cpu, 64)
		return strconv.Itoa(int(a * b))
	}
}

// mem 超售计算
func overmem(mem, multiple string) string {
	if multiple == "" {
		multiple = "1"
	}
	b, _ := strconv.ParseFloat(multiple, 64)

	// 内存单位可能为：空，Ki，Mi，Gi，Ti，Pi，Ei
	if strings.HasSuffix(mem, "Ki") {
		a, err := strconv.ParseFloat(strings.Trim(mem, "Ki"), 64)
		if err != nil {
			klog.Error("--------内存超售计算失败----------")
			klog.Error(err)
			return "1"
		}
		return strconv.Itoa(int(a*b)) + "Ki"
	} else if strings.HasSuffix(mem, "Mi") {
		a, err := strconv.ParseFloat(strings.Trim(mem, "Mi"), 64)
		if err != nil {
			klog.Error("--------内存超售计算失败----------")
			klog.Error(err)
			return "1"
		}
		return strconv.Itoa(int(a*b)) + "Mi"
	} else if strings.HasSuffix(mem, "Gi") {
		a, err := strconv.ParseFloat(strings.Trim(mem, "Gi"), 64)
		if err != nil {
			klog.Error("--------内存超售计算失败----------")
			klog.Error(err)
			return "1"
		}
		return strconv.Itoa(int(a*b)) + "Gi"
	} else if strings.HasSuffix(mem, "Ti") {
		a, err := strconv.ParseFloat(strings.Trim(mem, "Ti"), 64)
		if err != nil {
			klog.Error("--------内存超售计算失败----------")
			klog.Error(err)
			return "1"
		}
		return strconv.Itoa(int(a*b)) + "Ti"
	} else if strings.HasSuffix(mem, "Pi") {
		a, err := strconv.ParseFloat(strings.Trim(mem, "Pi"), 64)
		if err != nil {
			klog.Error("--------内存超售计算失败----------")
			klog.Error(err)
			return "1"
		}
		return strconv.Itoa(int(a*b)) + "Pi"
	} else if strings.HasSuffix(mem, "Ei") {
		a, err := strconv.ParseFloat(strings.Trim(mem, "Ei"), 64)
		if err != nil {
			klog.Error("--------内存超售计算失败----------")
			klog.Error(err)
			return "1"
		}
		return strconv.Itoa(int(a*b)) + "Ei"
	} else {
		// 单位为空时，单位默认为Bytes
		a, err := strconv.ParseFloat(mem, 64)
		if err != nil {
			klog.Error("--------内存超售计算失败----------")
			klog.Error(err)
			return "1"
		}
		return strconv.Itoa(int(a * b))
	}
}

////实际使用率计算
//func cpuutil(name string) int {
//	body :=conntroller.PromPost("sum(kube_pod_container_resource_requests_cpu_cores{node=~\""+ name+"\"})/sum (rate (container_cpu_usage_seconds_total{instance=~\""+name +"\"}[2m]))")
//	cpu:=new(conntroller.AutoGenerated)
//	json.Unmarshal(body,&cpu)
//	for _,v :=range cpu.Data.Result{
//		a,_ :=v.Value[1].(float32)
//		fmt.Println(a)
//		if a <1 {
//			return 1
//		}else if a > 1.5 && a < 2 {
//				return int(a) + 1
//			} else if a >=2{
//				return int(a)
//		}
//	}
//	return 1
//}
//
//func memutil(name string) int {
//	body :=conntroller.PromPost("sum(kube_pod_container_resource_requests_memory_bytes{node=~\""+ name+"\"})/sum (rate (container_memory_working_set_bytes{instance=~\""+name +"\"}[2m]))")
//	cpu:=new(conntroller.AutoGenerated)
//	json.Unmarshal(body,&cpu)
//	for _,v :=range cpu.Data.Result{
//		a,_ :=v.Value[1].(int)
//		fmt.Println(a)
//		if a >=2 {
//			return a
//		}
//	}
//	return 1
//
//}
