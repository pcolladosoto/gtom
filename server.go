package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type server struct {
	sync.RWMutex

	ctx    context.Context
	events []AnnotationResponse
}

func newServer() *server {
	return &server{ctx: context.Background()}
}

// root exists so that jsonds can be successfully added as a Grafana Data Source.
//
// If this exists then Grafana emits this when adding the datasource:
//
//	Success
//	Data source is working
//
// otherwise it emits "Unknown error"
func (s *server) root(w http.ResponseWriter, r *http.Request) {
	log.Printf("request for %v [%v]\n", r.URL.Path, r.Method)
	fmt.Fprintf(w, "ok\n")
}

func (s *server) search(w http.ResponseWriter, r *http.Request) {
	log.Printf("request for %v [%v]\n", r.URL.Path, r.Method)
	switch r.Method {
	case http.MethodPost:
		// Note this request is empty in new Grafana versions
		sr := SearchReq{}
		if err := json.NewDecoder(r.Body).Decode(&sr); err != nil {
			http.Error(w, fmt.Sprintf("json decode failure: %v", err), http.StatusBadRequest)
			return
		}

		if err := json.NewEncoder(w).Encode(SearchResponse{"foo", "faa"}); err != nil {
			http.Error(w, fmt.Sprintf("json decode failure: %v", err), http.StatusBadRequest)
		}
		return
	default:
		http.Error(w, "bad method; supported [POST]\n", http.StatusBadRequest)
	}
}

func (s *server) query(w http.ResponseWriter, r *http.Request) {
	log.Printf("request for %v [%v]\n", r.URL.Path, r.Method)
	switch r.Method {
	case http.MethodPost:
		qr := QueryReq{}
		if err := json.NewDecoder(r.Body).Decode(&qr); err != nil {
			http.Error(w, fmt.Sprintf("json decode failure: %v", err), http.StatusBadRequest)
			return
		}
		log.Printf("received query request: %+v\n", qr)
		w.Header().Set("Content-Type", "application/json")
		for _, target := range qr.Targets {
			if err := json.NewEncoder(w).Encode([]QueryResponseTimeserie{{
				Target: target.Target,
				Datapoints: []Datapoint{
					{0.5, float64(time.Now().Unix() * 1000)},
				},
			}}); err != nil {
				http.Error(w, fmt.Sprintf("json decode failure: %v", err), http.StatusBadRequest)
			}
		}
		// http.Error(w, "working on it...\n", http.StatusNotImplemented)
		return
	default:
		http.Error(w, "bad method; supported [POST]\n", http.StatusBadRequest)
	}
}

func (s *server) annotations(w http.ResponseWriter, r *http.Request) {
	log.Printf("request for %v [%v]\n", r.URL.Path, r.Method)
	switch r.Method {
	case http.MethodOptions:
	case http.MethodPost:
		ar := AnnotationsReq{}
		if err := json.NewDecoder(r.Body).Decode(&ar); err != nil {
			http.Error(w, fmt.Sprintf("json decode failure: %v", err), http.StatusBadRequest)
			return
		}

		evs := s.filterEvents(ar.Annotation, ar.Range.From, ar.Range.To)
		if err := json.NewEncoder(w).Encode(evs); err != nil {
			log.Printf("json enc: %+v", err)
		}
	default:
		http.Error(w, "bad method; supported OPTIONS, POST", http.StatusBadRequest)
		return
	}
}

func (s *server) filterEvents(a Annotation, from, to time.Time) []AnnotationResponse {
	events := []AnnotationResponse{}
	for _, event := range s.events {
		event.Annotation = a
		event.Annotation.ShowLine = true
		if event.Time > from.Unix()*1000 && event.Time < to.Unix()*1000 {
			events = append(events, event)
		}
	}
	return events
}
