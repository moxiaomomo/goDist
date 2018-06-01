package proxy

import "time"

type context struct {
}

func (c *context) StartAt() time.Time {
	return time.Now()
}

func (c *context) EndAt() time.Time {
	return time.Now()
}

func (c *context) SetAttr(key string, value interface{}) {

}

func (c *context) GetAttr(key string) interface{} {
	return nil
}
