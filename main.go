package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"time"
)

func main() {
	target := os.Args[1]
	deployment := os.Args[2]

	cmd := exec.Command("bosh", "-e", target, "-d", deployment, "vms", "--vitals", "--json")
	j, err := cmd.Output()
	if err != nil {
		panic("Could not run command: " + err.Error())
	}

	type response struct {
		Tables []struct {
			Rows []struct {
				Instance  string `json:"instance"`
				CreatedAt string `json:"vm_created_at"`
			} `json:"rows"`
		} `json:"tables"`
	}

	r := &response{}

	err = json.Unmarshal(j, &r)
	if err != nil {
		panic("Could not unmarshal json: " + err.Error())
	}

	type intermediate struct {
		Instance  string
		CreatedAt int64
	}

	inters := make([]intermediate, 0, len(r.Tables[0].Rows))

	for _, v := range r.Tables[0].Rows {
		t, err := time.Parse(time.UnixDate, v.CreatedAt)
		if err != nil {
			panic(fmt.Sprintf("Could not parse date `%s': %s", v.CreatedAt, err))
		}
		inters = append(inters, intermediate{
			Instance:  v.Instance,
			CreatedAt: t.Unix(),
		})
	}

	sort.Slice(inters, func(i, j int) bool { return inters[i].CreatedAt < inters[j].CreatedAt })

	type final struct {
		VM        string `json:"vm"`
		CreatedAt string `json:"created_at"`
	}

	finals := make([]final, 0, len(inters))

	for _, v := range inters {
		finals = append(finals, final{
			VM:        v.Instance,
			CreatedAt: time.Unix(v.CreatedAt, 0).Format(time.UnixDate),
		})
	}

	j, err = json.Marshal(&finals)
	if err != nil {
		panic("Could not marshal into json: " + err.Error())
	}

	jInd := &bytes.Buffer{}
	json.Indent(jInd, j, "", "  ")
	fmt.Println(jInd.String())
}
