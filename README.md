
<h2><a id="user-content-whats-tron" class="anchor" aria-hidden="true" href="#whats-tron"><svg class="octicon octicon-link" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path fill-rule="evenodd" d="M4 9h1v1H4c-1.5 0-3-1.69-3-3.5S2.55 3 4 3h4c1.45 0 3 1.69 3 3.5 0 1.41-.91 2.72-2 3.25V8.59c.58-.45 1-1.27 1-2.09C10 5.22 8.98 4 8 4H4c-.98 0-2 1.22-2 2.5S3 9 4 9zm9-3h-1v1h1c1 0 2 1.22 2 2.5S13.98 12 13 12H9c-.98 0-2-1.22-2-2.5 0-.83.42-1.64 1-2.09V6.25c-1.09.53-2 1.84-2 3.25C6 11.31 7.55 13 9 13h4c1.45 0 3-1.69 3-3.5S14.5 6 13 6z"></path></svg></a>What is shlog</h2>
<p>Shlog is simple logging package with log file rotation and severity level based log</p>
<p>Package publicly exposes six main functions for configuring logging system and initializing main parameters</p>

<p><strong>InitLogger()</strong> simply returns logger instance with default parameters if no other configuration is provided. With default parameters logging is
synchronous meaning every log request is blocking call and directly writes log data into file. Severity level is "Trace" (so every severity level log will be logged into file). By default there is no log rotation system so on every application startup, logging continues (or creates log file if it doesn't exist) into current working directory. </p>

<p><strong>Release()</strong> is called on application shut down which releases file descriptor (closes file) and flushes any log data currently residing into buffer (buffer is used in asynchronous mode logging). Please note that Release MUST be called during application shutdown (probably with defering)</p>

<p><strong>SetLogType(logType int)</strong> sets log type async/sync. Asynchronous logging uses buffering for execution time optimization in which
case log data might not be logged in file in timely fashion (mode used for production mode). Synchronous logging writes log data into file instantly and can be monitored in real time (mode used in development mode)</p>

<p><strong>SetLogLevel(logLevel int)</strong> sets log level constraint. There are five severity levels. Setting level means
that logs with severity level strictly less than level we set will not be logged (setting level is mainly used in
production mode when we don't need low level log to be logged)</p>

<p><strong>SetLogDirectory(logDirectory string)</strong> checks whether given string is a valid directory with appropriate permissions
and stores directory where log files will finally be created for logging.</p>

<p><strong>SetLogRotation(rotatedLogDir string, rotationType int)</strong> initializes log file rotation. first parameter must be directory with appropriate permissions for log files to be moved after rotation. second parameter is rotation type SizeBasedRotation rotates files when they reach 'defaultRotationSize' (1mb), TimeBasedRotation will rotate log file when day changes (meaning current date is different than file logging start date)
there are no other options for file rotations (not yet :))</p>

<p><strong>Log(level int, data string)</strong> is main logging function that takes severity level for this log
and simply log string to be stored in log file</p>


<h2><a id="user-content-prepare-dependencies" class="anchor" aria-hidden="true" href="#prepare-dependencies"><svg class="octicon octicon-link" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path fill-rule="evenodd" d="M4 9h1v1H4c-1.5 0-3-1.69-3-3.5S2.55 3 4 3h4c1.45 0 3 1.69 3 3.5 0 1.41-.91 2.72-2 3.25V8.59c.58-.45 1-1.27 1-2.09C10 5.22 8.98 4 8 4H4c-.98 0-2 1.22-2 2.5S3 9 4 9zm9-3h-1v1h1c1 0 2 1.22 2 2.5S13.98 12 13 12H9c-.98 0-2-1.22-2-2.5 0-.83.42-1.64 1-2.09V6.25c-1.09.53-2 1.84-2 3.25C6 11.31 7.55 13 9 13h4c1.45 0 3-1.69 3-3.5S14.5 6 13 6z"></path></svg></a>Example</h2>

<div class="highlight highlight-source-shell"><pre>
package main

import (
&nbsp;&nbsp;&nbsp;&nbsp;  "shlog"
&nbsp;&nbsp;&nbsp;&nbsp;  "time"
)


func main() {
&nbsp;&nbsp;&nbsp;&nbsp;  logger := shlog.InitLogger()

&nbsp;&nbsp;&nbsp;&nbsp;  //EXAMPLE: logger.SetLogRotation("src/rotation/file/directory", shlog.SizeBasedRotation)
&nbsp;&nbsp;&nbsp;&nbsp;  logger.SetLogRotation("src/rotation/file/directory", shlog.TimeBasedRotation)

&nbsp;&nbsp;&nbsp;&nbsp;  //EXAMPLE: logger.SetLogType(shlog.Asynchronous)
&nbsp;&nbsp;&nbsp;&nbsp;  logger.SetLogType(shlog.Synchronous)

&nbsp;&nbsp;&nbsp;&nbsp;  //EXAMPLE: logger.SetLogLevel(shlog.Warning)
&nbsp;&nbsp;&nbsp;&nbsp;  logger.SetLogLevel(shlog.Info)

&nbsp;&nbsp;&nbsp;&nbsp;  //EXAMPLE: by default log goes into current working directory
&nbsp;&nbsp;&nbsp;&nbsp;  logger.SetLogDirectory("/src/log/file/directory")

&nbsp;&nbsp;&nbsp;&nbsp;  //MUST BE CALLED
&nbsp;&nbsp;&nbsp;&nbsp;  defer logger.Release()



&nbsp;&nbsp;&nbsp;&nbsp;  //just log stuff
&nbsp;&nbsp;&nbsp;&nbsp;  for {
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;     time.Sleep(2 * time.Second)
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;     logger.Log(shlog.Trace,"Tracing this")
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;     logger.Log(shlog.Info,"Informing that ")
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;     logger.Log(shlog.Warning,"Warning those ")
&nbsp;&nbsp;&nbsp;&nbsp;   }

}

</pre></div>
