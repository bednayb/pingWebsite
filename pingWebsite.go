package pingWebsite

import (
	"crypto/tls"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

type pingResult struct {
	status string
	err    error
}

func PingTest(endpoint string, pingTimes int, maxSecPingingTime int) pingResult {

	var pingDataContainer []int64
	finishPingingBeforeTimer := make(chan pingResult)

	//timer
	go func() {
		time.Sleep(time.Second * time.Duration(maxSecPingingTime))
		finishPingingBeforeTimer <- pingResult{"[WARNING-SLOW]", nil}
	}()

	// make ping on website
	go func() {
		for i := 0; i < pingTimes; i++ {
			err := ping(endpoint, &pingDataContainer)
			if err != nil {
				finishPingingBeforeTimer <- pingResult{"[ERROR]", err}
				break
			}
		}
		finishPingingBeforeTimer <- pingResult{"[OK]", nil}
	}()

	pingStatus := <-finishPingingBeforeTimer
	close(finishPingingBeforeTimer)

	// errorHandling
	if pingStatus.status == "[ERROR]" {
		return pingResult{"[ERROR]", pingStatus.err}
	}

	if pingStatus.status == "[WARNING-SLOW]" {
		fmt.Println("[WARNING] pinging was longer than " + strconv.Itoa(maxSecPingingTime) + " second")
		return pingResult{"[WARNING] pinging was longer than " + strconv.Itoa(maxSecPingingTime) + " second", nil}
	}
	// most important information from data
	fmt.Println("average ping time", average(pingDataContainer), " millisecond")
	fmt.Println("deviation of ping time", deviation(pingDataContainer), " millisecond")
	fmt.Println("longest ping time", maximum(pingDataContainer), " millisecond")

	return pingResult{"[OK]", nil}
}

func ping(endpoint string, pingDataContainer *[]int64) error {

	// timer start
	t1 := time.Now()
	unixNano := t1.UnixNano()
	umillisec := unixNano / 1000000

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if nil != err {
		log.Print("Unable to create request: %v", err)
		return err
	}
	req.Close = true

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	/* Send it off */
	res, err := client.Do(req)
	if nil != err {
		log.Printf("Unable to request: %v", err)
		return err
	}
	if http.StatusOK != res.StatusCode {
		log.Printf("Ping Status: %v", res.Status)
		return err
	}

	res.Body.Close()

	// check timer
	t2 := time.Now()
	unixNano2 := t2.UnixNano()
	umillisec2 := unixNano2 / 1000000
	diff := umillisec2 - umillisec

	//saved how long was the pinging
	*pingDataContainer = append(*pingDataContainer, diff)

	return nil
}

func average(pingDataContainer []int64) int {
	var total int64
	for _, value := range pingDataContainer {
		total += value
	}

	x := total / int64(len(pingDataContainer))
	result := int(x)
	return result
}

func deviation(pingDataContainer []int64) string {
	avg := average(pingDataContainer)
	var deviation float64

	for _, v := range pingDataContainer {
		deviation += math.Pow(float64(v)-float64(avg), 2)
	}

	deviation = math.Sqrt(deviation / float64(len(pingDataContainer)))
	s := strconv.FormatFloat(deviation, 'f', 2, 64)

	return s
}

func maximum(pingDataContainer []int64) int {
	var maxPingTime int64

	for _, v := range pingDataContainer {
		if v > maxPingTime {
			maxPingTime = v
		}
	}
	return int(maxPingTime)
}
