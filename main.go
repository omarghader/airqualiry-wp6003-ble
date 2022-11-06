package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/currantlabs/ble"
	"github.com/currantlabs/ble/linux"
)

var (
	addr      = flag.String("addr", "", "address of remote peripheral (MAC on Linux, UUID on OS X)")
	bind      = flag.String("bind", ":6161", "the address to bind the http server")
	calibrate = flag.Bool("calibrate", false, "to calibrate the box")
	charUUID  = ble.Reverse([]byte{0xff, 0xf4})
	client    ble.Client
	profile   *ble.Profile

	temperature float32
	co2         int32
	tvoc        float32
	hcho        float32
)

func main() {
	flag.Parse()
	if *addr == "" {
		panic(errors.New("argument -addr should not be empty"))
	}
	connectBLE()

	var gracefulStop = make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		sig := <-gracefulStop
		fmt.Printf("caught sig: %+v\n", sig)
		disconnect()
		os.Exit(0)
	}()

	go func() {
		<-client.Disconnected()
		fmt.Printf("[ %s ] is disconnected \n", client.Address())
		panic("stop because disconnecting")
	}()

	// if Calibrate, send command and exit
	if *calibrate {
		Calibrate()
		return
		// exit here
	}

	if u := profile.Find(ble.NewCharacteristic(charUUID)); u != nil {
		err := client.Subscribe(u.(*ble.Characteristic), false, notificationHandler)
		if err != nil {
			fmt.Printf("can't subscribe to characteristic %s\n", err)
		}
	}

	hndle13UID, _ := ble.Parse("2902")
	hndl13 := profile.Find(ble.NewDescriptor(hndle13UID))
	handle10UID, _ := ble.Parse("fff1")
	hndl10 := profile.Find(ble.NewCharacteristic(handle10UID))

	client.WriteDescriptor(hndl13.(*ble.Descriptor), []byte{0x01, 0x00})
	client.WriteCharacteristic(hndl10.(*ble.Characteristic), []byte{0xaa, 0x15, 0x03, 0x1a, 0x00, 0x09, 0x35}, true)
	client.WriteCharacteristic(hndl10.(*ble.Characteristic), []byte{0xab}, true)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		res := map[string]interface{}{
			"temperature": temperature,
			"co2":         co2,
			"tvoc":        tvoc,
			"hcho":        hcho,
		}
		resJson, err := json.Marshal(res)
		if err != nil {
			http.Error(w, "Cannot marshal json.", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, string(resJson)+"\n")
	})

	fmt.Printf("Server started at port %s", *bind)
	http.ListenAndServe(*bind, nil)

}

func connectBLE() {
	d, err := linux.NewDevice()
	if err != nil {
		panic(err)
	}
	ble.SetDefaultDevice(d)

	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), time.Duration(5*time.Second)))
	addr := ble.NewAddr(*addr)

	fmt.Printf("Dialing to specified address: %s\n", addr)

	client, err = ble.Dial(ctx, addr)
	if err != nil {
		panic(err)
	}

	profile, err = client.DiscoverProfile(true)
	if err != nil {
		panic(err)
	}
}

func disconnect() {
	client.Unsubscribe(profile.Find(ble.NewCharacteristic(charUUID)).(*ble.Characteristic), false)
	client.CancelConnection()
}

func notificationHandler(req []byte) {
	if len(req) == 18 {
		temperature = float32(int32(req[6])<<8+int32(req[7])) / 10
		tvoc = float32(int32(req[10])<<8+int32(req[11])) / 1000
		hcho = float32(int32(req[12])<<8+int32(req[13])) / 1000
		co2 = int32(req[16])<<8 + int32(req[17]) - 150
		log.Printf("%f %f %f %d\n", temperature, tvoc, hcho, co2)
	}
}

func Calibrate() {
	handleSensorWrite, _ := ble.Parse("fff1")
	hndl10 := profile.Find(ble.NewCharacteristic(handleSensorWrite))

	client.WriteCharacteristic(hndl10.(*ble.Characteristic), []byte{0xad}, true)
}
