/*
Cliente
*/
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"util"
)

type Resp struct {
	Ok    bool   // true -> correcto, false -> error
	Msg   string // mensaje adicional
	Token []byte // token de sesión para utilizar por el cliente
}

// chk comprueba y sale si hay errores (ahorra escritura en programas sencillos)
func chk(e error) {
	if e != nil {
		panic(e)
	}
}

// Run gestiona el modo cliente
func main() {

	/* creamos un cliente especial que no comprueba la validez de los certificados
	esto es necesario por que usamos certificados autofirmados (para pruebas) */
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	cliUser := "usuario"
	cliPass := "contraseña"

	// derivación de claves sencilla a partir de los datos del usuario (así se evita enviar la contraseña en claro)
	keyClient := sha512.Sum512([]byte(cliUser + cliPass)) // un valor de 512 bits del que derivar claves
	keyLogin := keyClient[:32]                            // una mitad para el login (256 bits)
	keyData := keyClient[32:64]                           // la otra para los datos privados del cliente (256 bits)

	//datos del usuario (en este caso un par de claves pública/privada de RSA como ejemplo)
	cliPriv, err := rsa.GenerateKey(rand.Reader, 1024)
	chk(err)
	cliPriv.Precompute()       // aceleramos su uso con un precálculo
	cliPub := cliPriv.Public() // extraemos la clave pública por separado

	/*
		ejemplo de registro
	*/
	fmt.Print("\n\nREGISTRO\n\n")
	data := url.Values{}                      // estructura para contener los valores de la petición
	data.Set("cmd", "register")               // comando (string)
	data.Set("user", "usuario")               // usuario (string)
	data.Set("pass", util.Encode64(keyLogin)) // "contraseña" a base64 al ser []bytes (binario)

	// clave pública
	bytesPub, err := x509.MarshalPKIXPublicKey(cliPub) // serializamos con x509 (similar a JSON pero para especial para claves públicas)
	chk(err)
	compPub := util.Compress(bytesPub)         // comprimimos
	data.Set("pubkey", util.Encode64(compPub)) // a base64 sin cifrar (es pública)

	// clave privada
	bytesPriv, err := x509.MarshalPKCS8PrivateKey(cliPriv) // serializamos con x509 (similar a JSON pero especial para claves privadas)
	chk(err)
	compPriv := util.Compress(bytesPriv)       // comprimimos
	encPriv := util.Encrypt(compPriv, keyData) // ciframos con la clave de datos privados
	data.Set("prikey", util.Encode64(encPriv)) // a base64

	r, err := client.PostForm("https://localhost:10443", data) // enviamos por POST
	chk(err)
	io.Copy(os.Stdout, r.Body) // mostramos el cuerpo de la respuesta (es un reader)
	r.Body.Close()             // hay que cerrar el reader del body
	fmt.Println()

	/*
		ejemplo de login
	*/
	fmt.Print("\n\nLOGIN\n\n")
	data = url.Values{}
	data.Set("cmd", "login")                                  // comando (string)
	data.Set("user", "usuario")                               // usuario (string)
	data.Set("pass", util.Encode64(keyLogin))                 // contraseña (a base64 porque es []byte)
	r, err = client.PostForm("https://localhost:10443", data) // enviamos por POST
	chk(err)
	resp := Resp{}
	json.NewDecoder(r.Body).Decode(&resp) // decodificamos la respuesta para utilizar sus campos más adelante
	fmt.Println(resp)                     // imprimimos por pantalla
	r.Body.Close()                        // hay que cerrar el reader del body

	/*
		ejemplo de data sin utilizar el token correcto
	*/
	fmt.Print("\n\nDATA (BAD TOKEN)\n\n")
	badToken := make([]byte, 16)
	_, err = rand.Read(badToken)
	chk(err)

	data = url.Values{}
	data.Set("cmd", "data")                    // comando (string)
	data.Set("user", "usuario")                // usuario (string)
	data.Set("pass", util.Encode64(keyLogin))  // contraseña (a base64 porque es []byte)
	data.Set("token", util.Encode64(badToken)) // token incorrecto
	r, err = client.PostForm("https://localhost:10443", data)
	chk(err)

	io.Copy(os.Stdout, r.Body) // mostramos el cuerpo de la respuesta (es un reader)
	r.Body.Close()             // hay que cerrar el reader del body
	fmt.Println()

	/*
		ejemplo de data con token correcto
	*/
	fmt.Print("\n\nDATA (TOKEN OK)\n\n")
	data = url.Values{}
	data.Set("cmd", "data")                      // comando (string)
	data.Set("user", "usuario")                  // usuario (string)
	data.Set("pass", util.Encode64(keyLogin))    // contraseña (a base64 porque es []byte)
	data.Set("token", util.Encode64(resp.Token)) // token correcto
	r, err = client.PostForm("https://localhost:10443", data)
	chk(err)
	io.Copy(os.Stdout, r.Body) // mostramos el cuerpo de la respuesta (es un reader)
	r.Body.Close()             // hay que cerrar el reader del body
	fmt.Println()

}
