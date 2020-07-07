package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type StreamServer struct {}

func (s *StreamServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Access-Control-Allow-Origin", "*")

  session := strings.Split(r.URL.Path, "/")[0]

  now := time.Now().Unix()

  rows, err := db.Query(`SELECT sessions.sid, classes.room FROM sessions
  INNER JOIN people ON sessions.uname=people.uname
  INNER JOIN roster ON people.id=roster.pid
  INNER JOIN classes ON roster.cid=classes.id
  INNER JOIN periods ON classes.period=periods.code
  WHERE periods.stime<? AND periods.etime>? AND sessions.id=?;`, now, now, session)

  if err != nil {
    logger.Printf("Error querying sessions to stream! %s\n", err.Error())
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte("Error querying session to stream file!"))
    return
  }

  defer rows.Close()

  if !rows.Next() {
    logger.Printf("No current period found for stream!")
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(`{"status": false, "err": "No session for stream found"}`))
    return
  }

  var (
    sid string
    room uint64
  )

  err = rows.Scan(&sid, &room)

  fmt.Printf("About to serve %s/%s/%d\n", os.Getenv("FS_PATH"), sid, room)
  http.StripPrefix(session, http.FileServer(http.Dir(fmt.Sprintf("%s/%s/%d", os.Getenv("FS_PATH"), sid, room)))).ServeHTTP(w, r)
}

//func requestStream(w http.ResponseWriter, r *http.Request) {
//  w.Header().Set("Access-Control-Allow-Origin", "*")
//
//  query := r.URL.Query()
//
//  var session string
//  var sid string
//
//  if query["session"] == nil || query["sid"] == nil {
//    w.WriteHeader(http.StatusBadRequest)
//    w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
//    return
//  }
//
//  session = query["session"][0]
//  sid = query["sid"][0]
//
//  role, err := checkSession(sid, session)
//  if err != nil {
//    logger.Printf("Error in requestStream trying to check session! Error: %s\n", err.Error())
//    w.WriteHeader(http.StatusBadRequest)
//    w.Write([]byte(`{"status": false, "err": "Bad session!"}`))
//    return
//  }
//
//  if role == "A" {
//    scheduledSession := new(ScheduledSession)
//    scheduledSession.session = session
//    scheduledSession.streamFolder = sid
//    scheduledSession.endTime = 9223372036854775807
//    scheduledSessions = append(scheduledSessions, scheduledSession)
//    w.WriteHeader(http.StatusOK)
//    w.Write([]byte(fmt.Sprintf(`{"status": true, "err": ""}`)))
//    return
//  } else if role == "T" || role == "S" {
//    // Maybe split off teachers to admin like views?!?!?!
//    // SQL STRING: select * from sessions INNER JOIN people on sessions.uname = people.uname INNER JOIN roster ON people.id = roster.pid INNER JOIN classes ON roster.cid = classes.id INNER JOIN periods ON classes.period = periods.code;
//    for i, s := range scheduledSessions {
//      if s.session == session {
//        if s.endTime < uint64(time.Now().Unix()) {
//          scheduledSessions[len(scheduledSessions)-1], scheduledSessions[i] = scheduledSessions[i], scheduledSessions[len(scheduledSessions)-1]
//          scheduledSessions = scheduledSessions[:len(scheduledSessions)-1]
//          // Magic code to delete this scheduled session from the list of scheduledSessions
//          w.WriteHeader(http.StatusBadRequest)
//          w.Write([]byte(`{"status": false, "err": "Session too old"}`))
//          return
//        } else {
//          w.WriteHeader(http.StatusOK)
//          w.Write([]byte(`{"status": true, "err": false}`))
//          return
//        }
//      }
//    }
//
//    scheduledSession := new(ScheduledSession)
//    scheduledSession.session = session
//    rows, err := db.Query("SELECT classes.room, periods.code, periods.stime, periods.etime, classes.name, people.fname, people.lname FROM sessions INNER JOIN people ON sessions.uname = people.uname INNER JOIN roster ON people.id = roster.pid INNER JOIN classes ON roster.cid = classes.id INNER JOIN periods ON classes.period = periods.code WHERE sessions.sid=? AND sessions.id=?;", sid, session)
//    if err != nil {
//      logger.Printf("Error in requestStream trying to query database for student session! Error: %s\n", err.Error())
//      w.WriteHeader(http.StatusInternalServerError)
//      w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Failed to retrieve records for schedule query"}`)))
//      return
//    }
//    defer rows.Close()
//
//    for rows.Next() {
//      var (
//        className string
//        firstName string
//        lastName string
//        room string
//        period string
//        stime uint64
//        etime uint64
//      )
//
//      if err := rows.Scan(&room, &period, &stime, &etime, &className, &firstName, &lastName); err != nil {
//        logger.Printf("Error in requestStream trying to scan rows for data! Error: %s\n", err.Error())
//        w.WriteHeader(http.StatusInternalServerError)
//        w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Invalid data returned from scheduled query"}`)))
//        return
//      }
//
//      now := uint64(time.Now().Unix())
//
//      if now < etime && now > stime {
//        scheduledSession.streamFolder = fmt.Sprintf("%s/%s", sid, room)
//        scheduledSession.startTime = stime
//        scheduledSession.endTime = etime
//        scheduledSession.className = className
//        scheduledSession.firstName = firstName
//        scheduledSession.lastName = lastName
//        scheduledSessions = append(scheduledSessions, scheduledSession)
//        w.WriteHeader(http.StatusOK)
//        w.Write([]byte(fmt.Sprintf(`{"status": true, "err": ""}`)))
//        return
//      }
//    }
//
//    w.WriteHeader(http.StatusBadRequest)
//    w.Write([]byte(`{"status": false, "err": "No class found for current time"}`))
//  }
//}
//
