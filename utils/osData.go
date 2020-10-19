package utils

import "runtime"

const (
	OSKernel = "taco_os_kernel"
	OSFamily = "taco_os_family"
	Architecture = "taco_architecture"
)

type OSDataProvider struct {

}

func (odp OSDataProvider) GetTemplateVariables() map[string]interface{} {
	return map[string]interface{} {
		OSKernel: runtime.GOOS, //https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63#a-list-of-valid-goos-values
		Architecture: runtime.GOARCH, //https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63#a-list-of-valid-goarch-values
	}
}
