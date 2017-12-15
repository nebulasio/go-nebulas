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
  "syscall"
  "path/filepath"
  "time"
  "math/rand"
  "net"
  "strconv"
  "os/exec"
)

func InitCrashReporter(){
  os.Setenv("GOBACKTRACE", "crash")
  dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
  if err != nil {
    return ;
  }
  fp := fmt.Sprintf("%vcrash_%v.log", os.TempDir(), os.Getpid())

  fmt.Println(fp)
  port := rand.Intn(0xFFFF-1024) + 1024
  s, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

  for i:=0; i < 0xff; i+=1{
    if err != nil{
      port = rand.Intn(0xFFFF-1024) + 1024
      s, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
    }else{
      break;
    }
  }

  if err != nil{
    return ;
  }
  defer s.Close()


  code := rand.Intn(0xFFFF)
  cmd := exec.Command(fmt.Sprintf("%v/nebulas_crashreporter", dir),
  "-logfile",
  fp,
  "-port",
  strconv.Itoa(port),
  "-code",
  strconv.Itoa(code))

  cmd.Start()

  conn, err := s.Accept()
  if err != nil {
    return ;
  }
  var buf = make([]byte, 10)
  n, berror := conn.Read(buf);
  if berror != nil {
    //log.Println("conn read error:", berror)
    return
  }
  rs := string(buf[:n])

  if(rs == strconv.Itoa(code)){
    if crashFile, err := os.OpenFile(fp, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0664); err == nil {
      crashFile.WriteString(fmt.Sprintf("%v Opened crashfile at %v", os.Getpid(), time.Now()))
      os.Stderr = crashFile
      syscall.Dup2(int(crashFile.Fd()), 2)
    }
  }else{
    fmt.Println("code not match, recv:, should be ", rs, strconv.Itoa(code))
  }

}
