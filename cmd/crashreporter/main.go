// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

package main

import (
  "fmt"
  "os"
  "os/user"
  "time"
  "net"
  "strconv"
  "flag"
  "syscall"
  "io/ioutil"
  "strings"
  "path/filepath"
  "github.com/VividCortex/godaemon"
)

func checkCrashFileAndUpload(fp string) error{
  if _, ferr := os.Stat(fp); ferr == nil {
    bytes, err := ioutil.ReadFile(fp)
    if err != nil{
      return nil
    }
    lines := strings.Split(string(bytes), "\n")
    current, err := user.Current()
    for i, line := range lines {
      line = strings.Replace(line, current.HomeDir, "HomeDir", -1)
      line = strings.Replace(line, current.Name, "Name", -1)
      line = strings.Replace(line, current.Username, "Username", -1)
      lines[i] = line
    }
    output := strings.Join(lines, "\n")

    dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
    ioutil.WriteFile(fmt.Sprintf("%v/crash.log", dir), []byte(output), 0644)

    return nil
  }else{
    fmt.Println("no crash yet")
  }
  return nil
}

func main() {
  godaemon.MakeDaemon(&godaemon.DaemonAttr{})
  logfp :=flag.String("logfile","", "log file path")
  port :=flag.Int("port",0,"tcp port for notification")
  code :=flag.Int("code",0,"verification code")
  pid :=flag.Int("pid",0,"verification code")
  flag.Parse()
  s, err := net.Dial("tcp", fmt.Sprintf(":%d", *port))
  if err != nil{
    return ;
  }

  defer s.Close()

  s.Write([]byte(strconv.Itoa(*code)))

  ticker := time.NewTicker(1 * time.Second)
  for _ = range ticker.C {
    process, err := os.FindProcess(*pid)
    if err != nil {
      fmt.Printf("Failed to find process: %s\n", err)
      checkCrashFileAndUpload(*logfp)
      return ;
    } else {
      err := process.Signal(syscall.Signal(0))
      if err == nil{
        continue
      }else{
        checkCrashFileAndUpload(*logfp)
        return ;
      }
    }
  }
}
