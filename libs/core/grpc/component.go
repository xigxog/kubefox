package grpc

import (
	"fmt"
	"strings"
)

const Separator = ":"

func (c *Component) SetApp(val string) {
	c.App = val
}

func (c *Component) SetName(val string) {
	c.Name = val
}

func (c *Component) SetGitHash(val string) {
	c.GitHash = val
}

func (c *Component) SetId(val string) {
	c.Id = val
}

func (c *Component) GetURI() string {
	return strings.Join([]string{"kubefox:component", c.GetApp(), c.GetName(), c.GetGitHash(), c.GetId()}, Separator)
}

func (c *Component) GetKey() string {
	return strings.Join([]string{c.GetApp(), c.GetName()}, Separator)
}

func (c *Component) GetHTTPKey() string {
	return strings.Join([]string{c.GetApp(), c.GetName(), c.GetGitHash()}, Separator)
}

func (c *Component) GetRequestSubject() string {
	return fmt.Sprintf("%s.%s.req", c.GetName(), c.GetGitHash())
}

func (c *Component) GetResponseSubject() string {
	return fmt.Sprintf("%s.%s.%s.resp", c.GetName(), c.GetGitHash(), c.GetId())
}

func (c *Component) GetAsyncSubject() string {
	return fmt.Sprintf("%s.%s.async", c.GetName(), c.GetGitHash())
}

func (c *Component) GetRequestConsumer() string {
	return strings.ReplaceAll(c.GetRequestSubject(), ".", "_")
}

func (c *Component) GetResponseConsumer() string {
	return strings.ReplaceAll(c.GetResponseSubject(), ".", "_")
}

func (c *Component) GetAsyncConsumer() string {
	return strings.ReplaceAll(c.GetAsyncSubject(), ".", "_")
}

func (c *Component) GetStream() string {
	return fmt.Sprintf("%s_%s", strings.ToUpper(c.GetName()), c.GetGitHash())
}

func (c *Component) GetSubjectWildcard() string {
	return fmt.Sprintf("%s.%s.>", c.GetName(), c.GetGitHash())
}
