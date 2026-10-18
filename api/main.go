package main

import (
  "context"
  "encoding/json"
  "log"
  "net/http"
  "os"
  "strings"
  "sync"
  "time"
  "github.com/gin-gonic/gin"
  "github.com/jackc/pgx/v5/pgxpool"
  "github.com/redis/go-redis/v9"
)

type Job struct {
  ID string `json:"id"`
  ResourceType string `json:"resourceType"`
  Action string `json:"action"`
  Config map[string]any `json:"config"`
  Status string `json:"status"`
  Message string `json:"message"`
  CreatedAt time.Time `json:"createdAt"`
}
type Store struct {
  mu sync.RWMutex
  jobs map[string]Job
  events chan Job
  db *pgxpool.Pool
  cache *redis.Client
}
func env(k,d string) string { if v:=os.Getenv(k); v!="" { return v }; return d }
func allowed(xs []string, s string) bool { for _,x:=range xs { if x==s{return true} }; return false }

func main() {
  ctx:=context.Background()
  s:=&Store{jobs:map[string]Job{},events:make(chan Job,100)}
  if dsn:=os.Getenv("DATABASE_URL"); dsn!="" {
    p,e:=pgxpool.New(ctx,dsn)
    if e==nil { s.db=p; _,e=p.Exec(ctx,`CREATE TABLE IF NOT EXISTS jobs(id TEXT PRIMARY KEY, resource_type TEXT NOT NULL, action TEXT NOT NULL, config JSONB NOT NULL, status TEXT NOT NULL, message TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL)`) }
    if e!=nil { log.Printf("database initialization failed: %v",e); if p!=nil {p.Close()}; s.db=nil }
  }
  if a:=os.Getenv("REDIS_ADDR"); a!="" { s.cache=redis.NewClient(&redis.Options{Addr:a}) }
  go s.worker()
  r:=gin.New(); r.Use(gin.Logger(),gin.Recovery())
  r.GET("/health",func(c *gin.Context){c.JSON(200,gin.H{"status":"UP","queueMode":env("QUEUE_MODE","local")})})
  r.GET("/api/jobs",func(c *gin.Context){c.JSON(200,s.list())})
  r.POST("/api/jobs",func(c *gin.Context){
    var q struct { ResourceType string `json:"resourceType"`; Action string `json:"action"`; Config map[string]any `json:"config"` }
    if c.ShouldBindJSON(&q)!=nil { c.JSON(400,gin.H{"error":"invalid request body"}); return }
    q.ResourceType=strings.ToLower(strings.TrimSpace(q.ResourceType)); q.Action=strings.ToLower(strings.TrimSpace(q.Action))
    if !allowed([]string{"ec2","s3","iam","rds"},q.ResourceType)||!allowed([]string{"create","update","delete"},q.Action) { c.JSON(400,gin.H{"error":"unsupported resource type or action"}); return }
    j:=Job{ID:time.Now().UTC().Format("20060102T150405.000000000"),ResourceType:q.ResourceType,Action:q.Action,Config:q.Config,Status:"QUEUED",Message:"Request accepted",CreatedAt:time.Now().UTC()}
    if env("QUEUE_MODE","local")=="aws" { c.JSON(501,gin.H{"error":"SQS publisher is a documented integration placeholder; use local mode"}); return }
    s.save(j); select {case s.events<-j: default:}
    c.JSON(http.StatusAccepted,j)
  })
  r.GET("/api/events",func(c *gin.Context){
    c.Header("Content-Type","text/event-stream"); c.Header("Cache-Control","no-cache"); c.Header("Connection","keep-alive")
    t:=time.NewTicker(2*time.Second); defer t.Stop()
    for { select {
      case <-c.Request.Context().Done(): return
      case j:=<-s.events: b,_:=json.Marshal(j); c.SSEvent("job",string(b)); c.Writer.Flush()
      case <-t.C: b,_:=json.Marshal(s.list()); c.SSEvent("snapshot",string(b)); c.Writer.Flush()
    }}
  })
  _=r.Run(":"+env("API_PORT","8080"))
}
func(s *Store) save(j Job) {
  s.mu.Lock(); s.jobs[j.ID]=j; s.mu.Unlock()
  if s.db!=nil { b,_:=json.Marshal(j.Config); _,e:=s.db.Exec(context.Background(),`INSERT INTO jobs VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT(id) DO UPDATE SET status=EXCLUDED.status,message=EXCLUDED.message`,j.ID,j.ResourceType,j.Action,b,j.Status,j.Message,j.CreatedAt); if e!=nil {log.Printf("persist job: %v",e)} }
  if s.cache!=nil {_=s.cache.Set(context.Background(),"job:"+j.ID,j.Status,30*time.Minute).Err()}
}
func(s *Store) list() []Job { s.mu.RLock(); defer s.mu.RUnlock(); out:=make([]Job,0,len(s.jobs)); for _,j:=range s.jobs {out=append(out,j)}; return out }
func(s *Store) worker() { for j:=range s.events { if j.Status!="QUEUED"{continue}; time.Sleep(time.Second); j.Status="SUCCEEDED"; j.Message="Local simulation completed; no cloud resource created"; s.save(j) } }
