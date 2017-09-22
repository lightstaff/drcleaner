package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var log = logrus.New()

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
		cli.BoolFlag{
			Name: "verbose, V",
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
	vb := c.Bool("V")

	if vb {
		log.Info("start!!")
	}

	i := c.Args().First()
	u := c.String("U")
	ts := c.StringSlice("T")

	var client http.Client

	if 0 == len(ts) {
		tags, err := getTags(u, i)
		if err != nil {
			return cli.NewExitError(err.Error(), 86)
		}

		ts = tags.Tags
	}

	if vb {
		log.WithFields(logrus.Fields{
			"url":   u,
			"image": i,
			"tags":  strings.Join(ts, ","),
		}).Info("remove target")
	}

	for _, t := range ts {
		gURL := fmt.Sprintf("%s/v2/%s/manifests/%s", u, i, t)

		gReq, err := http.NewRequest("GET", gURL, nil)
		if err != nil {
			return cli.NewExitError(err.Error(), 86)
		}

		gReq.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")

		gRes, err := client.Do(gReq)
		if err != nil {
			return cli.NewExitError(err.Error(), 86)
		}
		defer gRes.Body.Close()

		if gRes.StatusCode < 200 && 300 < gRes.StatusCode {
			return cli.NewExitError(gRes.Status, 86)
		}

		r := gRes.Header.Get("Docker-Content-Digest")

		if vb {
			log.WithFields(logrus.Fields{
				"url":                   gURL,
				"status":                gRes.Status,
				"status_code":           gRes.StatusCode,
				"docker_content_digest": r,
			}).Info("get manifests")
		}

		dURL := fmt.Sprintf("%s/v2/%s/manifests/%s", u, i, r)

		dReq, err := http.NewRequest("DELETE", dURL, nil)
		if err != nil {
			return cli.NewExitError(err.Error(), 86)
		}

		dRes, err := client.Do(dReq)
		if err != nil {
			return cli.NewExitError(err.Error(), 86)
		}
		defer dRes.Body.Close()

		if dRes.StatusCode < 200 && 300 < dRes.StatusCode {
			return cli.NewExitError(dRes.Status, 86)
		}

		log.WithFields(logrus.Fields{
			"url":         dURL,
			"status":      dRes.Status,
			"status_code": dRes.StatusCode,
		}).Info("delete tag")
	}

	if vb {
		log.Info("finished!!")
	}

	return nil
}

func getTags(u, i string) (*tagsModel, error) {
	var client http.Client

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

	if res.StatusCode < 200 && 300 < res.StatusCode {
		return nil, errors.New(res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var v tagsModel

	if err := json.Unmarshal(body, &v); err != nil {
		return nil, err
	}

	return &v, nil
}
