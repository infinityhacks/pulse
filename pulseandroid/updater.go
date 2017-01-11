package pulseandroid

//Work in progress, Go version of the bash script

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/crypto/openpgp"
)

//Hardcoded 24F6D50F : Sajal Kayan (Signing key for TurboBytes Pulse) <sajal@turbobytes.com>
const pubkey = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1.4.12 (GNU/Linux)

mQMuBFV7Ar0RCADaT4Uzyq7ZHIkR5/WqbcLTkN9jxj4kp5XKzNQt9nrIbVw8nRt6
r3+2FFMPdLnSCqDpuz5X5pUnUlny+T5fgx0/OCJrz4J3iUgMftxc1TYN80rs5HuM
ClZqovw2T4VOvS+jRqJErzMUcAPIY4EPCNxQTWpcnjzQfrw5aLgAZ80wjZr7gpUf
dC2PgkW3QZtCtkTD8LB59fjeaVnRuWlQ7CXKX+MNxLGHD3BkZxHV7NBoc0TTJiHr
QGwS5/Ghiqbnm2julWmZKShB6s97ZDBfLCD4iSPbOZyKJIYlcGwhp3boqzL+714Q
2n16bZcEsnKI/Hle4tOKjJLk67rM7hM5oEdXAQCKAcvkNuAsTmEBg7PTa3iFfKxE
NDIS6A5r3qLWLISqwQf+NFJMa29AcTRSQNC587qjuxR/u2owBUtdkzyl0fIYeBXO
+LFTm9gJRXiNBsFI0A/qnyyAXHL4Vkf79hz6JW+jnFglvpXE0RebPSPLeWOdn3Bb
Mid8mm1iFagjstITqXy/RdzjFaoeTsl40JlyYiGPU2lvfMKWimVQ97E2Gn00kKrZ
HLvCHjANGY0nMnyUFroVdO9yZ3tM3dOFfL+TV/MnnaokFFOxbd7Gxq27ZcYIs7kO
alnroHHsWCCemidF0TAzJexF1AXAVfMMacxeJD3yPX6SUqPbloDf6WRPfAhjPIjw
WeRb3dhcd+/ct21gP5pG8U1pPJ+/yCGiVKn5MF8cwggA1xLmX1Xx4Z1Ncu+V6YHy
ZHbVAb3vtBZnL+hdYoJxpDoV7ML0SDX7ZsMXQ65eD0NHSCJehcK1jkYDwMvf8mi8
pQL3+veXsGh41uHPl9sFGHpZZCvfvggcDLr5Pa0gQuLOpUiXctUmw60B2Xvcp0js
6R98TKaeIyOJMVp3OTO95JaVZFxpYCJqzs5GFBroMpPYCIWn0vNLp3HOx2R59Y3Y
rfcD727Z0aG1MEnqWmShutTHXG/hm2no/nyDYxSWLq17ZQjhPO5pF6qcoy7zhhrh
6uJrfPLT41D2/HH4XDHKPBYxdyEBWt4EAC0bWcgSBnM8TcfdmfraFC+2DX9ZM7Q+
KbRFU2FqYWwgS2F5YW4gKFNpZ25pbmcga2V5IGZvciBUdXJib0J5dGVzIFB1bHNl
KSA8c2FqYWxAdHVyYm9ieXRlcy5jb20+iIEEExEIACkFAlV7Ar0CGwMFCQlmAYAH
CwkIBwMCAQYVCAIJCgsEFgIDAQIeAQIXgAAKCRC3o80eJPbVD/NbAQCJcBRISrWH
MC04vRPS/XLVTjJhLOApy0uMmfvbEZr6dAD9ESndQ71KPKQ3I/ikKJOEbBx9Kxzl
56OObA0/fiMHJns=
=x/VJ
-----END PGP PUBLIC KEY BLOCK-----`

var fingerprint = [20]byte{34, 135, 53, 47, 70, 86, 223, 73, 30, 2, 179, 209, 183, 163, 205, 30, 36, 246, 213, 15}

type Result struct {
	Error string
	Crt   string
}

type Updater struct {
	Current string
	bindir  string
	crtdir  string
	android AndroidService
}

//AndroidService is an interface created from Java that allows pulseandroid to do various tasks
// that need to be done from Java side
type AndroidService interface {
	GetConState() string       //Returns a string representation of connection type and signal strength.
	SensSMS(dest, body string) //Send an outgoing SMS to dest with payload body
	SendUSSN(dest string)      //Send USSN command
	SetStatus(status string)   //Set a status to show to user
	SetCSR(csr string)         //Set CSR to show in UI
}

func NewUpdater(bindir, crtdir string, android AndroidService) *Updater {
	u := &Updater{
		bindir:  bindir,
		crtdir:  crtdir,
		android: android,
	}
	u.android.SetStatus("Initializing")
	return u
}

//IngestSMS receives sender number and payload
func (u *Updater) IngestSMS(src, body string) {
	//TODO
}

//IngestUSSN receives incoming USSN message payload
func (u *Updater) IngestUSSN(body string) {
	//TODO
}

//IngestLog receives a line of log
func (u *Updater) IngestLog(line string) {
	//TODO
}

//LogDataUsage logs bytes used over network since last boot
func (u *Updater) LogDataUsage(rx, tx, total int64) {
	//TODO
	//Store the time and update and on subsiquent updates do a diff
}

func (u *Updater) Update() (result *Result) {
	u.android.SetStatus("Checking for updates")
	u.android.SetCSR("")
	result = &Result{}
	needsupdate := false
	//Check if we are on latest version...
	latest, err := getlatestversion()
	if err != nil {
		result.Error = err.Error()
		return
	}
	log.Println("Latest is", latest)
	current := u.getcurrentversion()
	log.Println("Current is", current)
	if latest != current {
		needsupdate = true
	}
	//Check if binary exists
	_, err = os.Stat(u.bindir + "/minion")
	if err != nil {
		needsupdate = true
	}

	if needsupdate {
		u.android.SetStatus("Updating")
		log.Println("Need to upgrade...")
		//Download the files...
		err = downloadfile("https://tb-minion.turbobytes.net/minion.android.arm.tar.gz", u.bindir+"/minion.android.arm.tar.gz")
		if err != nil {
			result.Error = err.Error()
			return
		}
		err = downloadfile("https://tb-minion.turbobytes.net/minion.android.arm.tar.gz.sig", u.bindir+"/minion.android.arm.tar.gz.sig")
		if err != nil {
			result.Error = err.Error()
			return
		}
		err = checkSig(u.bindir+"/minion.android.arm.tar.gz", u.bindir+"/minion.android.arm.tar.gz.sig")
		if err != nil {
			result.Error = err.Error()
			return
		}
		//Now untar...
		err = untardir(u.bindir, u.bindir+"/minion.android.arm.tar.gz")
		if err != nil {
			result.Error = err.Error()
			return
		}
		//Mark current
		writetxtfile(u.crtdir+"/current", latest)
	}
	result.Error = latest
	//Check for crt and update to my.tb if needed...
	log.Println("Checking crt things...")
	log.Println(fileexists(u.crtdir + "/minion.crt.request"))
	log.Println(fileexists(u.crtdir + "/minion.crt"))
	if fileexists(u.crtdir+"/minion.crt.request") && !fileexists(u.crtdir+"/minion.crt") {
		log.Println("crt missing")
		body, err := readtextfile(u.crtdir + "/minion.crt.request.tbid")
		log.Println("crt number exists")
		if err == nil {
			result.Crt = body
			u.android.SetStatus("CSR: " + body)
			u.android.SetCSR(body)
		} else {
			log.Println("crt number does not exists")
			body, err = readtextfile(u.crtdir + "/minion.crt.request")
			if err != nil {
				result.Error = err.Error()
				return
			}
			req, err := http.NewRequest("PUT", "https://my.turbobytes.com/pulse/upload_csr/", strings.NewReader(body))
			if err != nil {
				result.Error = err.Error()
				return
			}
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				result.Error = err.Error()
				return
			}
			if resp.StatusCode != 200 {
				result.Error = "Posting to my.tb " + resp.Status
				return
			}
			defer resp.Body.Close()
			d, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				result.Error = err.Error()
				return
			}
			log.Println(string(d))
			err = writetxtfile(u.crtdir+"/minion.crt.request.tbid", string(d))
			if err != nil {
				result.Error = err.Error()
				return
			}
			result.Crt = string(d)
			u.android.SetStatus("CSR: " + string(d))
			u.android.SetCSR(string(d))
		}
	}
	return
}

//Quick and dirty
func fileexists(fname string) bool {
	log.Println(fname)
	_, err := os.Stat(fname)
	if err != nil {
		return false
	}
	return true
}

func (u *Updater) ConStateExample() {
	for {
		st := time.Now()
		log.Println(u.android.GetConState())
		log.Println("GetConState() took:", time.Since(st))
		time.Sleep(time.Minute)
	}
}

func (u *Updater) Exec(cnc string) string {
	go u.ConStateExample()
	u.android.SetStatus("Launching agent binary to " + cnc)
	log.Println("BIN", u.bindir+"/minion")
	cmd := exec.Command(u.bindir+"/minion", "-cnc", cnc)
	cmd.Dir = u.crtdir
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err.Error()

	}
	// start the command after having set up the pipe
	if err := cmd.Start(); err != nil {
		return err.Error()
	}
	in := bufio.NewScanner(stderr)
	for in.Scan() {
		log.Printf(in.Text()) // write each line to your log, or anything you need
	}
	if err := in.Err(); err != nil {
		return err.Error()
	}
	return ""
}

func writetxtfile(fname, body string) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(body))
	return err
}

//getcurrentversion returns blank string when file missing
func (u *Updater) getcurrentversion() string {
	fpath := u.crtdir + "/current"
	strbody, err := readtextfile(fpath)
	if err != nil {
		return ""
	}
	return strbody
}

func readtextfile(fname string) (string, error) {
	f, err := os.Open(fname)
	if err != nil {
		return "", err
	}
	defer f.Close()
	body, err := ioutil.ReadAll(f)
	if err == nil {
		strbody := strings.TrimSpace(string(body))
		return strbody, nil
	}
	return "", err
}

func getlatestversion() (string, error) {
	resp, err := http.Get("https://tb-minion.turbobytes.net/latest")
	if err == nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			strbody := strings.TrimSpace(string(body))
			return strbody, nil
		} else {
			return "", err
		}
	} else {
		return "", err
	}
}

func downloadfile(url, dst string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(resp.Status + " : " + url)
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

//https://gist.github.com/indraniel/1a91458984179ab4cf80
func untardir(dir, fname string) error {
	log.Println("untardir", fname)
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	gzf, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(gzf)
	log.Println(tarReader)
	for {
		header, err := tarReader.Next()
		log.Println(header, err)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		name := header.Name
		if header.Typeflag == tar.TypeReg {
			log.Println(name)
			//Write the files in current dir
			f, err := os.Create(dir + "/" + name)
			if err != nil {
				return err
			}
			defer f.Close()
			io.Copy(f, tarReader)
			err = f.Chmod(0744)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

//https://gist.github.com/lsowen/d420a64821414cd2adfb
func checkSig(fileName string, sigFileName string) error {
	//Build keyring
	log.Println(fileName, sigFileName)
	keyRingReader := bytes.NewBufferString(pubkey)
	signature, err := os.Open(sigFileName)
	if err != nil {
		return err
		log.Println("sigfile", err)
	}
	defer signature.Close()

	verification_target, err := os.Open(fileName)
	if err != nil {
		return err
		log.Println("target", err)
	}
	defer verification_target.Close()
	keyring, err := openpgp.ReadArmoredKeyRing(keyRingReader)
	if err != nil {
		log.Println("ReadArmoredKeyRing", err)
		return err
	}
	log.Println(keyring)
	entity, err := openpgp.CheckDetachedSignature(keyring, verification_target, signature)
	if err != nil {
		log.Println("CheckDetachedSignature", err)
		return err
	}
	//Recheck
	if entity.PrimaryKey.Fingerprint != fingerprint {
		return errors.New("Fingerpring did not match")
	}
	return nil
}
