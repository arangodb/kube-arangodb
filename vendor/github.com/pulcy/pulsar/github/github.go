package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/juju/errgo"
	logging "github.com/op/go-logging"
)

var (
	maskAny = errgo.MaskFunc(errgo.Any)
)

type GithubService struct {
	Logger     *logging.Logger
	User       string // Github username
	Repository string // Github repository name
	Token      string // Github access token
}

type UploadAssetOptions struct {
	TagName  string // Name of release tag to attach asset to
	FileName string // Name of file
	Label    string // Label of file (optional)
	Path     string // Local path of file
}

func (s GithubService) ValidateCredentials() error {
	if s.User == "" {
		return maskAny(fmt.Errorf("User must be set"))
	}
	if s.Repository == "" {
		return maskAny(fmt.Errorf("Repository must be set"))
	}
	if s.Token == "" {
		return maskAny(fmt.Errorf("Token must be set"))
	}
	return nil
}

func (s GithubService) CreateRelease(options ReleaseCreate) error {
	if options.TagName == "" {
		return maskAny(fmt.Errorf("TagName must be set"))
	}
	options.Name = nvls(options.Name, options.TagName)
	options.Body = nvls(options.Body, options.TagName)

	if err := s.ValidateCredentials(); err != nil {
		return maskAny(err)
	}

	payload, err := json.Marshal(options)
	if err != nil {
		return maskAny(fmt.Errorf("can't encode release creation params, %v", err))
	}
	reader := bytes.NewReader(payload)

	uri := fmt.Sprintf("/repos/%s/%s/releases", s.User, s.Repository)
	resp, err := DoAuthRequest(s.Logger, "POST", ApiURL()+uri, "application/json", s.Token, nil, reader)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return maskAny(fmt.Errorf("while submitting %v, %v", string(payload), err))
	}

	s.Logger.Debugf("RESPONSE: %v", resp)
	if resp.StatusCode != http.StatusCreated {
		if resp.StatusCode == 422 {
			return maskAny(fmt.Errorf("github returned %v (this is probably because the release already exists)", resp.Status))
		}
		return maskAny(fmt.Errorf("github returned %v", resp.Status))
	}

	if s.Logger.IsEnabledFor(logging.DEBUG) {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return maskAny(fmt.Errorf("error while reading response, %v", err))
		}
		s.Logger.Debugf("BODY: %s", string(body))
	}

	return nil
}

func (s GithubService) UploadAsset(opt UploadAssetOptions) error {
	if opt.TagName == "" {
		return maskAny(fmt.Errorf("TagName must be set"))
	}
	if opt.FileName == "" {
		return maskAny(fmt.Errorf("FileName must be set"))
	}
	if opt.Path == "" {
		return maskAny(fmt.Errorf("Path must be set"))
	}
	if err := s.ValidateCredentials(); err != nil {
		return maskAny(err)
	}

	file, err := os.Open(opt.Path)
	if err != nil {
		return maskAny(err)
	}
	defer file.Close()

	/* find the release corresponding to the entered tag, if any */
	rel, err := ReleaseOfTag(s.Logger, s.User, s.Repository, opt.TagName, s.Token)
	if err != nil {
		return err
	}

	v := url.Values{}
	v.Set("name", opt.FileName)
	if opt.Label != "" {
		v.Set("label", opt.Label)
	}

	url := rel.CleanUploadUrl() + "?" + v.Encode()
	resp, err := DoAuthRequest(s.Logger, "POST", url, "application/octet-stream", s.Token, nil, file)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return maskAny(fmt.Errorf("can't create upload request to %v, %v", url, err))
	}

	s.Logger.Debugf("RESPONSE: %v", resp)
	if resp.StatusCode != http.StatusCreated {
		if msg, err := ToMessage(resp.Body); err == nil {
			return maskAny(fmt.Errorf("could not upload, status code (%v), %v", resp.Status, msg))
		} else {
			return maskAny(fmt.Errorf("could not upload, status code (%v)", resp.Status))
		}
	}

	if s.Logger.IsEnabledFor(logging.DEBUG) {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return maskAny(fmt.Errorf("error while reading response, %v", err))
		}
		s.Logger.Debugf("BODY: %s", string(body))
	}

	return nil
}
