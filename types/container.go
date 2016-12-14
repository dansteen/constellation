// types stores objects of various types and the functions to interact with them
package types

// Container stores all the information about a container to operate on
type Container struct {
	Name            string
	Image           string
	Environment     []string
	StateConditions []StateCondition
	Mounts          []Mount
	DependsOn       []*Container
}
