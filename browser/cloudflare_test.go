package browser

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"testing"
	"time"
)

// test old format
func TestCF01(t *testing.T) {
	testCFrequest(t, "testdata/js_challenge_before_11_12_2019.html", "GET", "168.5181599891")
}

func TestCF02(t *testing.T) {
	testCFrequest(t, "testdata/js_challenge_11_12_2019.html", "POST", "2.4164645335")
}

func testCFrequest(t *testing.T, page string, method string, jschl_answer string) {
	cfpage, err := ioutil.ReadFile(page)
	if err != nil {
		t.Errorf("error: %s", err)
		return
	}

	ts0 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "cloudflare-nginx")
		if os.Getenv("SURF_DEBUG_HEADERS") != "" {
			d, _ := httputil.DumpRequest(r,true)
			fmt.Printf("\n\n\n------ SERVER DUMP REQUEST ------\nRemote: %s\n%s\n------- END SERVER DUMP -------\n\n",r.RemoteAddr,d)
		}
		switch r.URL.Path {
		case "/feed":
			// got chellenge reponse as POST
			if r.Method == "POST" {
				cookie, err := r.Cookie("__cfduid")
				if err != nil {
					t.Fatalf("read cookie error: %s", err)
				}
				if cookie.Value != "d393e5efb2f6388ade72b3f1e0bc90bda1584672737" {
					t.Fatalf("no or wrong cookie set")
				}

				// check jschl_answer
				if r.FormValue("jschl_answer") == jschl_answer {
					w.WriteHeader(200)
				} else {
					t.Fatalf("wrong jschl_answer: got %s but expected %s", r.FormValue("jschl_answer"), jschl_answer)
				}
				// get final request after challenge with cookie set
			} else if r.Method == "GET" && len(r.Cookies()) == 1 {
				cookie,err := r.Cookie("__cfduid")
				if err != nil {
					t.Fatalf("read cookie error: %s",err)
				}
				if cookie.Value != "d393e5efb2f6388ade72b3f1e0bc90bda1584672737" {
					t.Fatalf("no or wrong cookie set")
				}
				w.WriteHeader(200)
				io.WriteString(w, "OK")

				// get initial GET request with no cookie set
			} else {
				if len(r.Cookies()) == 0 {
					http.SetCookie(w, &http.Cookie{
						Name:    "__cfduid",
						Value:   "d393e5efb2f6388ade72b3f1e0bc90bda1584672737",
						Expires: time.Now().AddDate(0, 0, 1),
					})
				} else {
					t.Fatalf("initial query has already cookie set?!")
				}
				w.WriteHeader(503)
				io.WriteString(w, string(cfpage))
			}

		// get chellenge response get request
		case "/cdn-cgi/l/chk_jschl":
			// check cookie is set
			cookie, err := r.Cookie("__cfduid")
			if err != nil {
				t.Fatalf("read cookie error: %s",err)
			}
			// check cookie value
			if cookie.Value != "d393e5efb2f6388ade72b3f1e0bc90bda1584672737" {
				t.Fatalf("wrong cookie set")
			}
			// check jschl_answer
			if r.FormValue("jschl_answer") == jschl_answer {
				w.WriteHeader(200)
			} else {
				t.Fatalf("wrong jschl_answer: got %s but expected %s",r.FormValue("jschl_answer"),jschl_answer)
			}
		default:

			//http.Error(w, "Unimplemented", 200)
			w.WriteHeader(404)
		}
	}))
	defer ts0.Close()

	// Alright, now let's see if the browser does the same thing
	bow := newDefaultTestBrowser()
	bow.UseCookie(true)
	if err := bow.Open(ts0.URL + "/feed?f=added%3A90d"); err != nil {
		t.Errorf("Failed to open url: %s", ts0.URL)
		return
	}

	if bow.StatusCode() != 200 {
		t.Fatalf("returned StatusCode is %d not 200",bow.StatusCode())
	}
}


