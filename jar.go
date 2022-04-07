package mycookiejar

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/bobesa/go-domain-util/domainutil"
)

const defaultFolder = "cookies"

type Jar struct {
	jar http.CookieJar
	sync.Mutex
	folder string
}

func New(jar http.CookieJar) Jar {
	return Jar{jar: jar, folder: defaultFolder}
}

func WithFolder(jar http.CookieJar, folder string) Jar {
	return Jar{jar: jar, folder: folder}
}

func (j *Jar) createFolder() error {
	if _, err := os.Stat(j.folder); os.IsNotExist(err) {
		err := os.MkdirAll(j.folder, os.FileMode(0755))
		if err != nil {
			return err
		}
	}
	return nil
}

func (j *Jar) save(host string, cookies []*http.Cookie) {
	j.Lock()
	defer j.Unlock()

	file := fmt.Sprintf("%v/%v.json", j.folder, host)

	dt, err := json.Marshal(&cookies)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(file, dt, os.FileMode(0655))
	if err != nil {
		log.Fatal(err)
	}
}

func (j *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	// save dulu ke jar
	j.jar.SetCookies(u, cookies)

	// lalu save ke file
	err := j.createFolder()
	if err != nil {
		log.Fatal(err)
	}

	host := u.Hostname()
	domain := domainutil.Domain(host)

	cks := j.jar.Cookies(u)

	// save host
	j.save(host, cks)

	// jika ada cookies domain
	if domain != "" && domain != host {
		j.save(domain, cookies)
	}
}
func (j *Jar) Cookies(u *url.URL) (cookies []*http.Cookie) {
	host := u.Hostname()
	domain := domainutil.Domain(host)

	// jika localhost
	if domain == "" {
		cks := j.parse(host)
		if len(cks) != 0 {
			j.jar.SetCookies(u, cks)
		}
		return j.jar.Cookies(u)
	}

	// jika domain tidak sama dengan host
	// maka host adalah subdomain
	// jadi parse cookies dari domain
	if domain != host {
		cks := j.parse(domain)
		if len(cks) != 0 {
			j.jar.SetCookies(u, cks)
		}
	}

	// parse host
	cks := j.parse(host)
	if len(cks) != 0 {
		j.jar.SetCookies(u, cks)
	}

	return j.jar.Cookies(u)

}

func (j *Jar) parse(host string) (cookies []*http.Cookie) {
	j.Lock()
	defer j.Unlock()

	file := fmt.Sprintf("%v/%v.json", j.folder, host)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil
	}

	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var cks []*http.Cookie
	json.NewDecoder(f).Decode(&cks)

	return cks
}
