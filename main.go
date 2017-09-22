package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var log = logrus.New()
var client http.Client

func main() {
	log.Out = os.Stdout

	app := cli.NewApp()
	app.Name = "drcleaner"
	app.Usage = "Docker Registry Cleaner"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "url, U",
			Value: "localhost:5000",
		},
		cli.StringSliceFlag{
			Name: "tags, T",
		},
	}

	app.Action = action

	app.Run(os.Args)
}

type tagsModel struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func action(c *cli.Context) error {
	if c.NArg() == 0 {
		err := errors.New("please input image name")
		log.Error(err.Error())
		return err
	}

	log.Info("docker registry cleaner start!!")

	i := c.Args().First()
	u := c.String("U")
	ts := c.StringSlice("T")

	if 0 == len(ts) {
		log.Infof("start get all tags from %s", i)

		tags, err := getAllTag(u, i)
		if err != nil {
			log.Error(err.Error())
			return err
		}

		ts = tags

		log.WithField("tags", tags).Infof("finish get all tags from %s", i)
	}

	for _, t := range ts {
		log.Infof("start get digest from %s:%s", i, t)

		d, err := getDigest(u, i, t)
		if err != nil {
			log.Error(err.Error())
			return err
		}

		log.WithField("digest", d).Infof("finish get digest from %s:%s", i, t)

		log.WithField("digest", d).Infof("start delete tag from %s:%s", i, t)

		if err := deleteTag(u, i, d); err != nil {
			log.Error(err.Error())
			return err
		}

		log.Infof("finish delete tag from %s:%s", i, t)
	}

	log.Info("docker registry cleaner finished!!")

	return nil
}

func getAllTag(u, i string) ([]string, error) {
	url := fmt.Sprintf("%s/v2/%s/tags/list", u, i)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || 300 < res.StatusCode {
		return nil, errors.New(res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var v map[string]interface{}

	if err := json.Unmarshal(body, &v); err != nil {
		return nil, err
	}

	tags, ok := v["tags"].([]string)
	if !ok {
		return nil, fmt.Errorf("%s is have not tag", i)
	}

	return tags, nil
}

func getDigest(u, i, t string) (string, error) {
	url := fmt.Sprintf("%s/v2/%s/manifests/%s", u, i, t)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || 300 < res.StatusCode {
		return "", errors.New(res.Status)
	}

	return res.Header.Get("Docker-Content-Digest"), nil
}

func deleteTag(u, i, d string) error {
	url := fmt.Sprintf("%s/v2/%s/manifests/%s", u, i, d)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || 300 < res.StatusCode {
		return errors.New(res.Status)
	}

	return nil
}
