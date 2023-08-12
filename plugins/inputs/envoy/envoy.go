package envoy

import (
	"crypto/tls"
	"encoding/json"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io"
	"net/http"
	"strings"
	"time"
)

const endpoint = "/production.json?details=1"

type Config struct {
	Url   string `toml:"envoy_url"`
	Token string `toml:"bearer_token"`
}

type EnergyGrid struct {
	Production  Energy
	Consumption Energy
	Net         Energy
}

type Energy struct {
	Watt   int
	Phases [3]int
}

type envoyResponse struct {
	Production  []envoyEnergy `json:"production,omitempty"`
	Consumption []envoyEnergy `json:"consumption,omitempty"`
}

type envoyEnergy struct {
	Type            string       `json:"type"`
	MeasurementType string       `json:"measurementType"`
	WNow            float64      `json:"wNow"`
	Phases          []envoyPhase `json:"Lines"`
}

type envoyPhase struct {
	WNow float64 `json:"wNow"`
}

func init() {
	inputs.Add("envoy", func() telegraf.Input {
		return &Config{}
	})
}

const sampleConfig = `
  ## Config management url.
  # envoy_url = "http://envoy.local"
  # bearer_token = "eyJraWQiOi..."
`

func (e Config) SampleConfig() string {
	return sampleConfig
}

func (e Config) Gather(accumulator telegraf.Accumulator) error {
	energyGrid := EnergyGrid{}

	now := time.Now()
	tags := map[string]string{}
	fields := map[string]interface{}{
		"total":          nil,
		"p1_production":  nil,
		"p1_consumption": nil,
		"p1_net":         nil,
		"p2_production":  nil,
		"p2_consumption": nil,
		"p2_net":         nil,
		"p3_production":  nil,
		"p3_consumption": nil,
		"p3_net":         nil,
	}
	url := strings.TrimSuffix(e.Url, "/")
	//Envoy latest firmware forces redirect to https with unsigned certificate.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", url+endpoint, nil)
	if err != nil {
		return err
	}

	token := "Bearer " + e.Token
	req.Header.Set("Authorization", token)
	req.Header.Add("Accept", "application/json")
	httpResp, err := client.Do(req)

	if err != nil {
		//Envoy sometimes loses LAN connection, send nil values to reflect that.
		accumulator.AddFields("envoy", fields, tags, now)

		return nil
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(httpResp.Body)

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	request := envoyResponse{}
	_ = json.Unmarshal(body, &request)

	for _, prod := range request.Production {
		if prod.Type == "eim" {
			updateResponse(&energyGrid.Production, prod)
		}
	}

	for _, cons := range request.Consumption {
		if cons.MeasurementType == "total-consumption" {
			updateResponse(&energyGrid.Consumption, cons)
		}
		if cons.MeasurementType == "net-consumption" {
			updateResponse(&energyGrid.Net, cons)
		}
	}

	fields = map[string]interface{}{
		"total":          energyGrid.Net.Watt,
		"p1_production":  energyGrid.Production.Phases[0],
		"p1_consumption": energyGrid.Consumption.Phases[0],
		"p1_net":         energyGrid.Net.Phases[0],
		"p2_production":  energyGrid.Production.Phases[1],
		"p2_consumption": energyGrid.Consumption.Phases[1],
		"p2_net":         energyGrid.Net.Phases[1],
		"p3_production":  energyGrid.Production.Phases[2],
		"p3_consumption": energyGrid.Consumption.Phases[2],
		"p3_net":         energyGrid.Net.Phases[2],
	}

	accumulator.AddFields("envoy", fields, tags, now)

	return nil
}

func (e Config) Description() string {
	return "Gather consumption information from a three-phase envoyResponse."
}

func updateResponse(energyResponse *Energy, energyRequest envoyEnergy) {
	energyResponse.Watt = int(energyRequest.WNow)
	for key, phase := range energyRequest.Phases {
		energyResponse.Phases[key] = int(phase.WNow)
	}
}
