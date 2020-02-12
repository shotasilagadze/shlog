package main

import (
  "shlog"
  "time"
)


func main() {
  logger := shlog.InitLogger()

  //EXAMPLE: logger.SetLogRotation("src/rotation/file/directory", shlog.SizeBasedRotation)
  logger.SetLogRotation("src/rotation/file/directory", shlog.TimeBasedRotation)

  //EXAMPLE: logger.SetLogType(shlog.Asynchronous)
  logger.SetLogType(shlog.Synchronous)

  //EXAMPLE: logger.SetLogLevel(shlog.Warning)
  logger.SetLogLevel(shlog.Info)

  //EXAMPLE: by default log goes into current working directory
  logger.SetLogDirectory("/src/log/file/directory")

  //MUST BE CALLED
  defer logger.Release()



  //just log stuff
  for {
     time.Sleep(2 * time.Second)
     logger.Log(shlog.Trace,"Tracing this")
     logger.Log(shlog.Info,"Informing that ")
     logger.Log(shlog.Warning,"Warning those ")
   }



}
