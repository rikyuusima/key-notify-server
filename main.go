/* Package: サーバ本体
*/
package main

// import
import(
  "fmt"
  "github.com/equanz/key-notify-server/sql_query"
  "github.com/gin-gonic/gin"
  "encoding/json"
  "time"
  "net/http"
)

// app_id認証body用struct
type StatusJSON struct{
  AppID string `json:"app_id" binding:"required"`
}

// serverの作成
func main(){
  r := gin.Default()

  // root
  r.GET("/", func(c *gin.Context){
    c.Redirect(http.StatusMovedPermanently, "/app")
  })

  // file access
  r.StaticFS("/app", http.Dir("app"))

  // favicon access
  r.StaticFile("/favicon.ico", "./resources/favicon.ico")

  // '/api' group
  api := r.Group("/api")
  {
    /* key on or off
     * :status: string "on", "off"文字列をそれぞれハードウェアのon, offとして処理
    */
    api.POST("/hard/:status", func(c *gin.Context){
      var body StatusJSON
      c.BindJSON(&body) // bind post body
      state := c.Param("status")

      has_app_id, err := sql_query.Has_app_id(body.AppID) // search app_id

      if err == nil {
        if body.AppID != "" && has_app_id == true {
          if state == "on" {
            if send_form_message(true) == nil {
              if sql_query.Insert_status("ON") == nil {
                c.String(200, "ON")
              } else {
                c.String(400, "Bad Request")
              }
            } else {
              c.String(400, "Bad Request")
            }
          } else if state == "off" {
            if send_form_message(false) == nil {
              if sql_query.Insert_status("OFF") == nil {
                c.String(200, "OFF")
              } else {
                c.String(400, "Bad Request")
              }
            } else {
              c.String(400, "Bad Request")
            }
          } else {
            // error
            c.String(400, "Bad Request")
          }
        } else {
          // error
          c.String(400, "Bad Request")
        }
      } else {
        // error
        c.String(503, "Service Unavailable")
      }
    })

    /* on/off統計データ(rawデータ)を指定日時から指定日時まで返す
     * first_date: string(DATETIMEフォーマット) 統計値の取得を開始する指定日時
     * end_date: string(DATETIMEフォーマット) 統計値の取得を終わる指定日時
    */
    api.GET("/statistics", func(c *gin.Context){
      q := c.Request.URL.Query() // query params
      fd, fd_ok := q["first_date"]
      ed, ed_ok := q["end_date"]

      if fd_ok == true && ed_ok == true {
        first_time, err_first := time.Parse("2006-01-02 15:04:05", fd[0]) // parse to format
        if err_first != nil {
          fmt.Println(err_first)
          c.String(400, "Bad Request")
        } else {
          end_time, err_end := time.Parse("2006-01-02 15:04:05", ed[0]) // parse to format
          if err_end != nil {
            fmt.Println(err_end)
            c.String(400, "Bad Request")
          } else {
            info_array, err_sql := sql_query.Get_statistics(first_time, end_time)
            info_array_json, err_json := json.Marshal(info_array) // generate json bytes from struct
            if err_sql != nil {
              fmt.Println(err_sql)
              c.String(400, "Bad Request")
            } else if err_json != nil {
              fmt.Println(err_json)
              c.String(400, "Bad Request")
            } else {
              c.String(200, string(info_array_json)) // stringify and response
            }
          }
        }
      } else {
        info_array, err_sql := sql_query.Get_all_statistics()
        info_array_json, err_json := json.Marshal(info_array) // generate json bytes from struct
        if err_sql != nil {
          fmt.Println(err_sql)
          c.String(400, "Bad Request")
        } else if err_json != nil {
          fmt.Println(err_json)
          c.String(400, "Bad Request")
        } else {
          c.String(200, string(info_array_json)) // stringify and response
        }
      }
    })

    /* on/off統計データ(rawデータ)の最新1件を返す
    */
    api.GET("/statistic",func(c *gin.Context){
      info_array, err_sql := sql_query.Get_latest_state()
      info_array_json, err_json := json.Marshal(info_array)
      if err_sql != nil {
        fmt.Println(err_sql)
        c.String(400, "Bad Request")
      } else if err_json != nil {
        fmt.Println(err_json)
        c.String(400, "Bad Request")
      } else {
        c.String(200, string(info_array_json))
      }
    })

    /* on/off統計データ(rawデータ)の指定された日時より早いデータを1件返す
     * date: string(DATETIMEフォーマット) 取得する値の指定日時
    */
    api.GET("/before_statistic",func(c *gin.Context){
      q := c.Request.URL.Query() // query params
      fd, fd_ok := q["date"]

      if fd_ok == true {
        time, err := time.Parse("2006-01-02 15:04:05", fd[0]) // parse to format
        if err != nil {
          fmt.Println(err)
          c.String(400, "Bad Request")
        } else {
          info_array, err_sql := sql_query.Get_before_state(time)
          info_array_json, err_json := json.Marshal(info_array)
          if err_sql != nil {
            fmt.Println(err_sql)
            c.String(400, "Bad Request")
          } else if err_json != nil {
            fmt.Println(err_json)
            c.String(400, "Bad Request")
          } else {
            c.String(200, string(info_array_json))
          }
        }
      } else {
        c.String(400, "Bad Request")
      }
    })
  }

  r.Run()
}
