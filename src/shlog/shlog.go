package shlog
import (
    "bufio"
    "errors"
    "io/ioutil"
    "os"
    "path/filepath"
    "runtime"
    "strconv"
    "sync"
    "time"
)
//Main const variables used during log system initialization
const (
    //default buffer size in case writes are asynchronous and therefore buffered
    defaultBufSize      = 4096 * 2
    defaultRotationSize = 1048576
    //severity levels for logging
    Trace   = 0
    Info    = 1
    Warning = 2
    Error   = 3
    Fatal   = 4
    //log file rotation types
    SizeBasedRotation = 0
    TimeBasedRotation = 1
    //async or sync log
    Synchronous  = 0
    Asynchronous = 1
)
var (
    //severity level mismatch
    errSeverityLevel = errors.New("Severity level excluded")
    //severity level names used in log messages
    severityLevels = [5]string{"trace", "info", "warning", "error", "fatal"}
)
/*
    log structure stores main logging system parameters and appropriate values
    which are initialized during application startup
    NOTE: log attributes must be initialized before any log operation or behaviour will be undefined
*/
//LOG struct main parameters
type LOG struct {
    //file path
    logFilePath string
    //open file
    file *os.File
    //buffered writer in case of asynchronous log
    writer *bufio.Writer
    //rotation structure in case of file rotation is requested
    rotation *Rotation
    //severity constraint for this instance
    logLevel int
    // synchronous/asynchronous log
    logType int
    //synchronizer for write locking
    writeLocker *sync.Mutex
}

//Log file rotation parameters
type Rotation struct {
    //size-based/time-based
    rotationType int
    //where rotated log data files go after rotation
    rotatedLogDir string
    //RotationSize
    rotationSize int64
    //current file size in bytes
    currentFileSize int64
    //current file creation or logging start date
    currentFileDate string


}

/*
    InitLogger initializes LOG instance which is the only instance
    alive during application lifetime and stores main parameters such as
    log type (async/sync), log file rotation parameters, severity level
    constraints and some other primitives for thread safely
*/
func InitLogger() *LOG {
    //initialize LOG default values
    var log LOG
    log.logFilePath = "data.log"
    log.writer = nil
    log.rotation = nil
    log.logLevel = Trace
    log.writeLocker = &sync.Mutex{}
    log.logType = Synchronous
    return &log
}
/*
    Following four functions are used to initialize logging system
    in the beggining. it sets log type, severity level to which messages with
    low severity level are excluded from logging, logging directory and logging
    rotation system (including directory for rotated log files, rotation type)
    NOTE: all log attributes must be initialized before any log operation or behaviour will be undefined
*/


/*
SetLogType sets log type async/sync. Asynchronous logging uses buffering for execution time optimization in which
case log data might not be logged in file in timely fashion (mode used for production mode). Synchronous logging
writes log data into file instantly and can be monitored in real time (mode used in development mode)
*/
func (log *LOG) SetLogType(logType int) *LOG{
    log.logType = logType
    return log
}

/*
SetLogLevel sets log level constraint. There are five severity levels. Setting level means
that logs with severity level strictly less than level we set will not be logged (setting level is mainly used in
production mode when we don't need low level log to be logged)
*/
func (log *LOG) SetLogLevel(logLevel int) *LOG{
    log.logLevel = logLevel
    return log
}
/*
SetLogDirectory checks whether given string is a valid directory with appropriate permissions
and stores directory where log files will finally be created for logging.
*/
func (log *LOG) SetLogDirectory(logDirectory string) *LOG {
    if err := pathIsValid(&logDirectory); err != nil {
        panic(err)
    }
    log.logFilePath = logDirectory + "/data.log"
    return log
}


/*
SetLogRotation initializes log file rotation. first parameter must be directory with
appropriate permissions for log files to be moved after rotation. second parameter is rotation type
SizeBasedRotation rotates files when they reach 'defaultRotationSize', TimeBasedRotation will rotate log file
when day changes (meaning current date is different than file logging start date)
there are no other options for file rotations
*/
func (log *LOG) SetLogRotation(rotatedLogDir string, rotationType int) *LOG{
    //check rotation type is either file based or time based
    if rotationType != SizeBasedRotation && rotationType != TimeBasedRotation {
        panic(errors.New("Log file rotation type error"))
    }
    //check if logRotationDir exists and has write permission
    if err := pathIsValid(&rotatedLogDir); err != nil {
        panic(err)
    }
    //create default parameters for log file rotation
    log.rotation = &Rotation{rotationType: rotationType, rotationSize: defaultRotationSize, rotatedLogDir: rotatedLogDir}
    return log
}


/*
Release (NOTE) MUST be called during application shut down
it makes sure that buffered log data (in case of asynchronous logging)
will be flushed to the file and closes in the end
*/
func (log *LOG) Release() {
    if log.writer != nil {
        log.writer.Flush()
    }
    log.file.Close()
}




/*
Log exposes main logging function that takes severity level for this log
and simply log string to be stored in log file
*/
func (log *LOG) Log(level int, data string) error {

    //get time and file log parameters
    logTimePrefix := getLogTime()
    logFilePrefix := getLogFilePlace()
    logSeverityName := severityLevels[level]


    //synchronize writes
    log.writeLocker.Lock()
    defer log.writeLocker.Unlock()

    //check open file (double check pattern)
    if log.file == nil && log.writer == nil {
            err := log.openFile()
            if err != nil {
                panic(err)
            }
    }

    //check log levels
    if level < 0 || level > 4 || level < log.logLevel {
        return errSeverityLevel
    }


    var writtenByteNum int
    if log.writer == nil {
        writtenByteNum, _ = log.file.Write([]byte(logSeverityName + "  " + logTimePrefix + " - " + logFilePrefix + ":  " + data + "\n"))
    } else {
        writtenByteNum, _ = log.writer.WriteString(logSeverityName + "  " + logTimePrefix + " - " + logFilePrefix + ":  " + data + "\n")
    }
    //increase current file size
    log.rotation.currentFileSize = log.rotation.currentFileSize + int64(writtenByteNum)


    if log.rotation != nil && log.rotation.rotationType == SizeBasedRotation && log.rotation.currentFileSize >= log.rotation.rotationSize {
        //check if it's time to rotate current log file
        log.sizeBasedRotation(logTimePrefix)
    }

    if log.rotation != nil && log.rotation.rotationType == TimeBasedRotation && log.rotation.currentFileDate != getCurrentDateString() {
        //check if it's time to rotate current log file
        log.timeBasedRotation(logTimePrefix)
    }


    return nil
}




//function simply checks if directory path is in fact directory and we have write permission on it
func pathIsValid(path *string) error {
    //remove last slash
    if (*path)[len(*path)-1] == '/' {
        *path = (*path)[0 : len(*path)-1]
    }

    //check create directory if not exists
    err := os.MkdirAll(*path,0766)
    if err != nil {
        panic(err)
    }


    //check directory exists
    info, err := os.Stat(*path)
    if err != nil {
        return err
    }
    //if directory is file instead of dir return err
    if !info.IsDir() {
        return errors.New("Log rotation directory parameter not valid")
    }
    //check dir is writable by creating and removing file in there
    f := filepath.Join(*path, ".touch")
    if err := ioutil.WriteFile(f, []byte(""), 0600); err != nil {
        return err
    }
    os.Remove(f)
    return nil
}



/*
    getLogTime simply takes current time precisely and
    creates string date-time prefix for each log message
*/
func getLogTime() string {
    var logTime string
    now := time.Now()
    year, month, day := now.Date()
    logTime = logTime + strconv.Itoa(year) + " "
    logTime = logTime + month.String() + " "
    logTime = logTime + strconv.Itoa(day) + " "
    hour, min, sec := now.Clock()
    logTime = logTime + strconv.Itoa(hour) + ":"
    logTime = logTime + strconv.Itoa(min) + ":"
    logTime = logTime + strconv.Itoa(sec)
    return logTime
}
//getLogFile returns file name and line of log occurence
func getLogFilePlace() string {
    _, file, line, ok := runtime.Caller(2)
    if !ok {
        file = "???"
        line = 0
    }
    //get last file name
    for i := len(file) - 1; i > 0; i-- {
        if file[i] == '/' {
            file = file[i+1:]
            break
        }
    }
    return file + ":" + strconv.Itoa(line)
}


//openFile is called on first log write and opens file with appropriate permissions and parameters
func (log *LOG) openFile() error {
    //open file
    file, err := os.OpenFile(log.logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
    if err != nil {
        return err
    }
    //create buffered writer if asynchronous log is requested
    if log.logType == Asynchronous {
        log.writer = bufio.NewWriterSize(file, defaultBufSize)
    } else {
        log.file = file
    }

    fi, err := file.Stat()
    if err != nil {
        panic(err)
    }
    log.rotation.currentFileSize = fi.Size()
    log.rotation.currentFileDate = getCurrentDateString()


    return nil
}



//returns current date string
func getCurrentDateString() string {
  var curDate string
  now := time.Now()
  year, month, day := now.Date()
  curDate = curDate + strconv.Itoa(year) + " "
  curDate = curDate + month.String() + " "
  curDate = curDate + strconv.Itoa(day)

  // _, min, _ := now.Clock()
  // curDate = curDate + strconv.Itoa(min)
  return curDate
}


/*
SizeBasedRotation makes rotation by moving
current log file into logRotationDir and creating new log file for next log writes
*/
func (log *LOG) sizeBasedRotation(timePrefix string) {
    //flush last log data residing in buffer
    if log.writer != nil {
        log.writer.Flush()
    }

    //close current log file
    log.file.Close()

    // rename file and move to log rotation directory
    newFile := log.rotation.rotatedLogDir+"/"+timePrefix+"-data.log"
    err := os.Rename(log.logFilePath, newFile)
    if err != nil {
      panic(err)
    }

    //open new log file
    log.writer = nil
    log.openFile()

    return
}


/*
timeBasedRotation makes rotation by moving
current log file into logRotationDir and creating new log file for next log writes
*/
func (log *LOG) timeBasedRotation(timePrefix string) {

    //flush last log data residing in buffer
    if log.writer != nil {
        log.writer.Flush()
    }

    //close current log file
    log.file.Close()

    // rename file and move to log rotation directory
    newFile := log.rotation.rotatedLogDir+"/"+timePrefix+"-data.log"
    err := os.Rename(log.logFilePath, newFile)
    if err != nil {
      panic(err)
    }

    //open new log file
    log.writer = nil
    log.openFile()

    //update current file size
    log.rotation.currentFileSize = 0
    log.rotation.currentFileDate = getCurrentDateString()

    return
}
