package wireguard

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

var Host = "127.0.0.1"
var Port = ":9991"
var Input = "Input from my side, пока, £, 语汉"
var InputFromOtherSide = "Input from other side, пока, £, 语汉, 123"

func TestWireguard(t *testing.T) {
	configA, err := FromWgQuick(`
[Interface]
PrivateKey = 2OZeP9sbnTBiyn1+43610zdMHhhE3CpaBJFxRJl5gGI=
Address = 10.0.0.1
ListenPort = 43234

[Peer]
PublicKey = fw2pUc5mHyrSLe43NG+Rb90isqFKnKmK2Et0Ma76CkY=
AllowedIPs = 10.0.0.2/32
`, "tunnelA")
	assert.NoError(t, err)

	configB, err := FromWgQuick(`
[Interface]
PrivateKey = kBXqMKQPlxmJPuxCxsmd+xuoQxZQocKlI2w1sB8zFnI=
Address = 10.0.0.2

[Peer]
PublicKey = h761vZ6TghHSmFuuEsAXRMJj8WLHkGhfyQXLcaXS2Xs=
AllowedIPs = 10.0.0.1/32
Endpoint = localhost:43234
`, "tunnelB")
	assert.NoError(t, err)

	tunnelA, err := CreateTunnel(configA)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	tunnelB, err := CreateTunnel(configB)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	listener, err := tunnelA.Listen("tcp", ":43235")
	assert.NoError(t, err)

	var received bytes.Buffer
	listenDone := make(chan struct{})

	go func() {
		conn, err := listener.Accept()
		assert.NoError(t, err)
		defer conn.Close()
		defer close(listenDone)

		_, err = io.Copy(&received, conn)
		assert.NoError(t, err)
	}()

	conn, err := tunnelB.Dial("tcp", "10.0.0.1:43235")
	assert.NoError(t, err)

	input := bytes.NewBufferString("Test string")
	_, err = io.Copy(conn, input)
	assert.NoError(t, err)
	conn.Close()

	// wait for listen to finish
	<-listenDone

	assert.Equal(t, "Test string", received.String())
}
