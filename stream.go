package main

import (
	"fmt"
	"net/http"
	"time"
)

type ScheduledSession struct {
  session string
  streamFolder string
  endTime int64
}

var scheduledSessions []*ScheduledSession

type StreamServer struct {}

func (s *StreamServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Access-Control-Allow-Origin", "*")

  query := r.URL.Query()

  var roomFolder string

  if query["session"] == nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(`{"status": false, "err": "Missing session"}`))
    return
  }

  for i, s := range scheduledSessions {
    if s.session == query["session"][0] {
      if s.endTime < time.Now().Unix() {
        scheduledSessions[len(scheduledSessions)-1], scheduledSessions[i] = scheduledSessions[i], scheduledSessions[len(scheduledSessions)-1]
        scheduledSessions = scheduledSessions[:len(scheduledSessions)-1]
        // Magic code to delete this scheduled session from the list of scheduledSessions
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte(`{"status": false, "err": "Session expired"}`))
        return
      } else {
        roomFolder = s.streamFolder
        break
      }
    }
  }

  http.FileServer(http.Dir(fmt.Sprintf("./streams/%s", roomFolder))).ServeHTTP(w, r)
}

func requestStream(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Access-Control-Allow-Origin", "*")

  query := r.URL.Query()

  var session string

  if query["session"] == nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(`{"status": false, "err": "Missing session"}`))
    return
  }

  session = query["session"][0]

  role, _ := checkSession(session)

  if role == "A" {
    scheduledSession := new(ScheduledSession)
    scheduledSession.session = session
    scheduledSession.streamFolder = ""
    scheduledSession.endTime = 9223372036854775807
    scheduledSessions = append(scheduledSessions, scheduledSession)
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(fmt.Sprintf(`{"status": true, "err": ""}`)))
    return
  } else if role == "T" || role == "S" {
    // Maybe split off teachers to admin like views?!?!?!
    // SQL STRING: select * from sessions INNER JOIN people on sessions.uname = people.uname INNER JOIN roster ON people.id = roster.pid INNER JOIN classes ON roster.cid = classes.id INNER JOIN periods ON classes.period = periods.code;

    scheduledSession := new(ScheduledSession)
    scheduledSession.session = session
    rows, err := db.Query("SELECT classes.room, periods.code, periods.stime, periods.etime, periods.date FROM sessions INNER JOIN people ON sessions.uname = people.uname INNER JOIN roster ON people.id = roster.pid INNER JOIN classes ON roster.cid = classes.id INNER JOIN periods ON classes.period = periods.code WHERE sessions.id=?;", session)
    if err != nil {
      fmt.Println(err.Error())
      w.WriteHeader(http.StatusInternalServerError)
      w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Failed to retrieve records for schedule query"}`)))
      return
    }
    defer rows.Close()

    for rows.Next() {
      var (
        room string
        period string
        stime uint64
        etime uint64
        date string
      )

      if err := rows.Scan(&room, &period, &stime, &etime, &date); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Invalid data returned from scheduled query"}`)))
        return
      }

      fmt.Println(date)

      scheduledSession.streamFolder = room

    }

    scheduledSession.streamFolder = ""
    scheduledSession.endTime = 9223372036854775807
    scheduledSessions = append(scheduledSessions, scheduledSession)
    w.WriteHeader(http.StatusOK)
  }
}

