package main

import(
  "fmt"
  "github.com/gocolly/colly"
  "encoding/json"
  "sync"
  "github.com/oleiade/lane"
  "os"
  "time"
  "log"
)


type metadata struct{
  Name string `json:"name"`
  Description string `json:"description"`
  Title string `json:"title"`
  Thumbnail string `json:"image"`
  Site string `json:"source"`
  Url string `json:url`
  VideoUrl string `json:"videourl"`
  VideoType string `json:"videotype"`
  Keywords string `json:"keywords"`
  Paid string `json:"paid"`
  ChannelId string `json:"channelid"`
  VideoId string `json:"videoid"`
  Duration string `json:"duration"`
  Genre string `json:"genre"`
  Unlisted string `json:"unlisted"`
  Width string `json:"width"`
  Height string `json:"height"`
  IsFamilyFriendly string `json:"isFamilyFriendly"`
  InteractionCount string `json:"interactionCount"`
  DatePublished string `json:"datePublished"`
}
 var finished bool

func main()  {
  finished = false
  var wgcrawler sync.WaitGroup
  var wgwriter sync.WaitGroup
  c := colly.NewCollector()
  c.AllowedDomains=[]string{
    "www.youtube.com",
  }
  c.AllowURLRevisit=false
  c.MaxDepth=1
  filename := fmt.Sprintf("video_%v.json",time.Now().Unix())
  fmt.Println(filename)
  file,err := os.OpenFile(filename,os.O_RDWR|os.O_CREATE,0755)
  if err != nil{
    log.Fatal(err)
  }
  _,err = file.Write([]byte("["))
  if err != nil{
    log.Fatal(err)
  }
  defer file.Close()
  var ch = make(chan string,10)
  queue := lane.NewQueue()
  queue.Enqueue("https://www.youtube.com/watch?v=kpmXrIo8Ncs")
  wgwriter.Add(1)
  go writer(ch,&wgwriter,file)
  for queue.Head() != nil{
    wgcrawler.Add(1)
    url := queue.Dequeue()
    go crawler(c.Clone(),&wgcrawler,url.(string),queue,ch)
    time.Sleep(3*time.Second)
  }
  wgcrawler.Wait()
  finished=true
  wgwriter.Wait()
    _,err = file.Write([]byte("]"))
    if err != nil{
      log.Fatal(err)
    }
    file.Sync()
}
func writer(ch chan string,wg *sync.WaitGroup, file* os.File)  {
  for {
      if !finished {
          data := <- ch
          file.Write([]byte(data))
      }else{
        fmt.Println("NO data !")
        break
      }
  }
  wg.Done()
}

func crawler(c *colly.Collector, wg *sync.WaitGroup, url string, queue *lane.Queue,ch chan string)  {
  var meta metadata
  c.OnHTML("a[href]",func(e *colly.HTMLElement)  {
      queue.Enqueue(e.Request.AbsoluteURL(e.Attr("href")))
  })
  c.OnHTML("meta[itemprop]",func(e *colly.HTMLElement)  {
      switch e.Attr("itemprop"){
      case "name":
        meta.Name = e.Attr("content")
      case "description":
        meta.Description = e.Attr("content")
      case "paid":
        meta.Paid = e.Attr("content")
      case "channelId":
        meta.ChannelId = e.Attr("content")
      case "videoId":
        meta.VideoId = e.Attr("content")
      case "duration":
        meta.Duration = e.Attr("content")
      case "unlisted":
        meta.Unlisted = e.Attr("content")
      case "width":
        meta.Width = e.Attr("content")
      case "height":
        meta.Height = e.Attr("content")
      case "isFamilyFriendly":
        meta.IsFamilyFriendly = e.Attr("content")
      case "interactionCount":
        meta.InteractionCount = e.Attr("content")
      case "datePublished":
        meta.DatePublished = e.Attr("content")
      case "genre":
        meta.Genre = e.Attr("content")
      }
  })
  c.OnHTML("meta[property]",func(e *colly.HTMLElement)  {
      switch e.Attr("property"){
      case "og:site_name":
        meta.Site = e.Attr("content")
      case "og:url":
        meta.Url = e.Attr("content")
      case "og:title":
        meta.Title = e.Attr("content")
      case "og:image":
        meta.Thumbnail = e.Attr("content")
      case "og:description":
        meta.Description = e.Attr("content")
      case "og:type":
        //meta.Name = e.Attr("content")
      case "og:video:url":
        meta.VideoUrl = e.Attr("content")
      case "og:video:type":
        meta.VideoType = e.Attr("content")
      case "og:video:tag":
      }
  })
  c.OnHTML("meta[name]",func(e *colly.HTMLElement)  {
      switch e.Attr("name"){
      case "title":
        meta.Title = e.Attr("content")
      case "description":
        meta.Description = e.Attr("content")
      case "keywords":
        meta.Name = e.Attr("content")
      }
  })
  c.OnRequest(func(r *colly.Request)  {
    fmt.Println("VISITIN URL->",r.URL)

  })
  c.OnResponse(func(r *colly.Response){
    //fmt.Println(string(r.Body))
  })
  c.Visit(url)
  if meta.VideoUrl != ""{
    data,_:=json.Marshal(meta)
    ch <- string(data)+","
  }
  wg.Done()
}
