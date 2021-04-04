package middleware

import (
	"github.com/robfig/cron/v3"
)

func StartTime(f func()){
	c := cron.New()
	spec := "*/5 * * * * ?"
  c.AddFunc(spec, f)
	c.Start()
}