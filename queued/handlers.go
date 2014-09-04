package queued

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func (s *Server) CreateQueueHandler(w http.ResponseWriter, req *http.Request) {
	config := NewQueueConfig()

	name := req.URL.Query().Get("name")
	if name == "" {
		send(w, http.StatusBadRequest, Json{"error": "required parameter is missing name"})
		return
	}
	config.DeadLetterQueue = req.URL.Query().Get("dead_letter_queue")

	if config.DeadLetterQueue != "" {
		config.Redirve = true
	}
	max, err := Atoi(req.URL.Query().Get("max_receives"))
	if err != nil {
		send(w, http.StatusBadRequest, Json{"error": "Invalid MaximumReceives parameter"})
		return
	}
	config.MaximumReceives = max

	timeout, err := strconv.Atoi(req.URL.Query().Get("timeout"))
	if err != nil {
		send(w, http.StatusBadRequest, Json{"error": "Invalid timeout parameter"})
		return
	}
	config.Timeout = timeout
	config.Name = name

	if req.URL.Query().Get("exponential_backoff") == "true" {
		config.ExponentialBackoff = true
	} else {
		config.ExponentialBackoff = false
	}

	queue, err := s.App.CreateQueue(config)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": err.Error()})
		return
	}
	bytes, err := json.Marshal(queue.config)

	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": err.Error()})
		return
	}

	w.Header().Set("Location", fmt.Sprintf("http://%s/%s", req.Host, name))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(bytes)
}

func (s *Server) EnqueueHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	value, err := ioutil.ReadAll(req.Body)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": err.Error()})
		return
	}

	mime := req.Header.Get("Content-Type")
	record, err := s.App.Enqueue(params["queue"], value, mime)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": err.Error()})
		return
	}

	w.Header().Set("Location", url(req, record))
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) DequeueHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	wait, err := Stod(req.URL.Query().Get("wait"), time.Second)
	if err != nil {
		send(w, http.StatusBadRequest, Json{"error": "Invalid wait parameter"})
		return
	}

	timeout, err := Stod(req.URL.Query().Get("timeout"), time.Second)
	if err != nil {
		send(w, http.StatusBadRequest, Json{"error": "Invalid timeout parameter"})
		return
	}

	record, err := s.App.Dequeue(params["queue"], wait, timeout)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": "Dequeue failed"})
		return
	}

	if record != nil {
		w.Header().Set("Location", url(req, record))
		w.Header().Set("Content-Type", record.ContentType())
		w.Write(record.Value)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *Server) InfoHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		send(w, http.StatusNotFound, Json{"error": "Item not found"})
		return
	}

	info, err := s.App.Info(params["queue"], id)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": "Failed to read item"})
		return
	}

	if info != nil {
		dequeued := "false"
		if info.dequeued {
			dequeued = "true"
		}

		w.Header().Set("X-Dequeued", dequeued)
		w.Header().Set("Content-Type", info.record.ContentType())
		w.Write(info.record.Value)
	} else {
		send(w, http.StatusNotFound, Json{"error": "Item not found"})
	}
}

func (s *Server) ListQueuesHandler(w http.ResponseWriter, req *http.Request) {
	queues := s.App.ListQueues()
	bytes, err := json.Marshal(queues)

	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)

}

func (s *Server) StatsHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	stats, err := s.App.Stats(params["queue"])
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": err.Error()})
		return
	}

	result := map[string]interface{}{}
	for field, value := range stats {
		result[field] = value
	}

	send(w, http.StatusOK, result)
}

func (s *Server) CompleteHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		send(w, http.StatusNotFound, Json{"error": "Item not found"})
		return
	}

	ok, err := s.App.Complete(params["queue"], id)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": "Complete failed"})
		return
	}

	if ok {
		w.WriteHeader(http.StatusNoContent)
	} else {
		send(w, http.StatusBadRequest, Json{"error": "Item not dequeued with timeout"})
	}
}

// Helpers

type Json map[string]interface{}

func Stod(val string, scale ...time.Duration) (time.Duration, error) {
	duration := NilDuration

	if val != "" {
		n, err := strconv.Atoi(val)

		if err != nil {
			return duration, err
		} else {
			duration = time.Duration(n)

			if len(scale) == 1 {
				duration *= scale[0]
			}
		}
	}

	return duration, nil
}

func Atoi(val string) (int, error) {
	if val != "" {
		return strconv.Atoi(val)
	}
	return 0, nil
}

func send(w http.ResponseWriter, code int, data Json) error {
	bytes, err := json.Marshal(data)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(bytes)

	return nil
}

func auth(config *Config, next http.Handler) http.Handler {
	unauthorized := func(w http.ResponseWriter) {
		send(w, http.StatusUnauthorized, Json{"error": "Unauthorized"})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if config.Auth != "" {
			s := strings.SplitN(req.Header.Get("Authorization"), " ", 2)
			if len(s) != 2 || s[0] != "Basic" {
				unauthorized(w)
				return
			}

			base, err := base64.StdEncoding.DecodeString(s[1])
			if err != nil {
				unauthorized(w)
				return
			}

			pair := strings.SplitN(string(base), ":", 2)
			if len(pair) != 2 {
				unauthorized(w)
				return
			}

			password := pair[1]
			if config.Auth != password {
				unauthorized(w)
				return
			}
		}

		next.ServeHTTP(w, req)
	})
}

func url(req *http.Request, record *Record) string {
	return fmt.Sprintf("http://%s/%s/%d", req.Host, record.Queue, record.Id)
}
