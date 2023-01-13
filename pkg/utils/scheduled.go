package utils

import (
        "errors"
        "strconv"
        "strings"
        "sync"
        "sync/atomic"
        "time"

        "github.com/sirupsen/logrus"
)

type TaskTime struct {
        Hour, Minute, Second int
}

func (tt TaskTime) IsOnTime() bool {
        return tt.Hour == time.Now().Hour() &&
                tt.Minute == time.Now().Minute() &&
                tt.Second == time.Now().Second()
}

func ParseTaskTime(s string) (TaskTime, error) {
        _default := TaskTime{Hour: 0, Minute: 0, Second: 0}

        parsed := strings.Split(s, ":")
        if len(parsed) < 3 || s == "" {
                return _default, errors.New("invalid format")
        }

        // hour
        hh, err := strconv.Atoi(parsed[0])
        if err != nil {
                return _default, errors.New("invalid hour")
        }

        // minute
        mm, err := strconv.Atoi(parsed[1])
        if err != nil {
                return _default, errors.New("invalid minute")
        }

        // second
        ss, err := strconv.Atoi(parsed[2])
        if err != nil {
                return _default, errors.New("invalid second")
        }

        return TaskTime{Hour: hh, Minute: mm, Second: ss}, nil
}

type Taskfunc func() error

type ScheduledTaskGroup struct {
        quit     chan bool
        nrunning int32 // number of running tasks
        logger   *logrus.Entry
        once     *sync.Once
}

func NewScheduledTaskGroup(namespace string) *ScheduledTaskGroup {
        logger := logrus.New()
        logger.SetFormatter(&logrus.TextFormatter{
                ForceColors:   false,
                DisableColors: false,
                FullTimestamp: true,
        })
        return &ScheduledTaskGroup{
                quit:     make(chan bool, 1),
                nrunning: 0,
                logger:   logger.WithField("taskgroup", namespace),
                once:     &sync.Once{},
        }
}

func (ptg *ScheduledTaskGroup) AddDaily(at TaskTime, fn Taskfunc) {
        ptg.taskAdded()
        go ptg.doAt(at, fn)
}

func (ptg *ScheduledTaskGroup) AddPerodical(interval time.Duration, fn Taskfunc) {
        ptg.taskAdded()
        go ptg.doEvery(interval, fn)
}

func (ptg *ScheduledTaskGroup) WaitAndStop() {
        ptg.once.Do(func() {
                // kill and wait all running tasks
                for i := int32(0); i < ptg.Running(); i++ {
                        ptg.quit <- true
                }
                ptg.waitTasksDone()

                close(ptg.quit)
        })

}

func (ptg *ScheduledTaskGroup) doEvery(interval time.Duration, fn Taskfunc) {
        defer ptg.taskDone()

        ticker := time.NewTicker(interval)
        defer ticker.Stop()

        for {
                select {
                case <-ptg.quit:
                        return

                case <-ticker.C:
                        err := retryTask(fn, 4)
                        if err != nil {
                                ptg.logger.Error(err)
                        }
                }
        }

}

func (ptg *ScheduledTaskGroup) doAt(at TaskTime, fn Taskfunc) {
        defer ptg.taskDone()

        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()

        for {
                select {
                case <-ptg.quit:
                        return

                case <-ticker.C:
                        if at.IsOnTime() {
                                err := retryTask(fn, 4)
                                if err != nil {
                                        ptg.logger.Error(err)
                                }
                        }

                }
        }
}

func (ptg *ScheduledTaskGroup) waitTasksDone() {
        nrunning := atomic.LoadInt32(&ptg.nrunning)

        for ; nrunning > 0; nrunning = atomic.LoadInt32(&ptg.nrunning) {
        }
}

func (ptg *ScheduledTaskGroup) taskAdded() {
        atomic.AddInt32(&ptg.nrunning, 1)
}

func (ptg *ScheduledTaskGroup) taskDone() {
        atomic.AddInt32(&ptg.nrunning, -1)
}

func (ptg *ScheduledTaskGroup) Running() int32 {
        return atomic.LoadInt32(&ptg.nrunning)
}

func retryTask(fn Taskfunc, retries int) (err error) {
retry_loop:
        for retry := 0; retry < retries; retry++ {
                err = fn()
                if err != nil {
                        time.Sleep(time.Duration(retry+1) * time.Second)
                        continue retry_loop
                } else {
                        break retry_loop
                }
        }

        return
}