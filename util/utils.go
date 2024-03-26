/*
Funciones comunes y dades
*/
package util

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"log/slog"
	"os"
)

func GetLogger() *slog.Logger {
	return slog.New(slog.Default().Handler())
}

// FailOnError comprueba y sale si hay errores (ahorra escritura en programas sencillos)
func FailOnError(e error) {
	if e != nil {
		panic(e)
	}
}

func DecodeJSON[T any](r io.Reader, v *T) error {
	dec := json.NewDecoder(r)
	return dec.Decode(v)
}

func EncodeJSON[T any](v T) []byte {
	buffer := new(bytes.Buffer)
	dec := json.NewEncoder(buffer)
	FailOnError(dec.Encode(&v))
	return buffer.Bytes()
}

// función para cifrar (AES-CTR 256), adjunta el IV al principio
func Encrypt(data, key []byte) (out []byte) {
	out = make([]byte, len(data)+16)    // reservamos espacio para el IV al principio
	rand.Read(out[:16])                 // generamos el IV
	blk, err := aes.NewCipher(key)      // cifrador en bloque (AES), usa key
	FailOnError(err)                    // comprobamos el error
	ctr := cipher.NewCTR(blk, out[:16]) // cifrador en flujo: modo CTR, usa IV
	ctr.XORKeyStream(out[16:], data)    // ciframos los datos
	return
}

// función para descifrar (AES-CTR 256)
func Decrypt(data, key []byte) (out []byte, err error) {
	out = make([]byte, len(data)-16) // la salida no va a tener el IV
	blk, err := aes.NewCipher(key)   // cifrador en bloque (AES), usa key
	if err != nil {
		return
	}
	// FailOnError(err)                     // comprobamos el error
	ctr := cipher.NewCTR(blk, data[:16]) // cifrador en flujo: modo CTR, usa IV
	ctr.XORKeyStream(out, data[16:])     // desciframos (doble cifrado) los datos
	return
}

// función para comprimir
func Compress(data []byte) []byte {
	var b bytes.Buffer      // b contendrá los datos comprimidos (tamaño variable)
	w := zlib.NewWriter(&b) // escritor que comprime sobre b
	w.Write(data)           // escribimos los datos
	w.Close()               // cerramos el escritor (buffering)
	return b.Bytes()        // devolvemos los datos comprimidos
}

// función para descomprimir
func Decompress(data []byte) []byte {
	var b bytes.Buffer // b contendrá los datos descomprimidos

	r, err := zlib.NewReader(bytes.NewReader(data)) // lector descomprime al leer

	FailOnError(err) // comprobamos el error
	io.Copy(&b, r)   // copiamos del descompresor (r) al buffer (b)
	r.Close()        // cerramos el lector (buffering)
	return b.Bytes() // devolvemos los datos descomprimidos
}

// función para codificar de []bytes a string (Base64)
func Encode64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data) // sólo za caracteres "imprimibles"
}

// función para decodificar de string a []bytes (Base64)
func Decode64(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s) // recupera el formato original
	FailOnError(err)                             // comprobamos el error
	return b                                     // devolvemos los datos originales
}

// función para resumir (SHA256)
func Hash(data []byte) []byte {
	h := sha256.New() // creamos un nuevo hash (SHA2-256)
	h.Write(data)     // procesamos los datos
	return h.Sum(nil) // obtenemos el resumen
}

func WriteECDSAKeyToFile(filename string, key *ecdsa.PrivateKey) {
	keyBytes, err := x509.MarshalECPrivateKey(key)
	FailOnError(err)

	privBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	}

	privFile, err := os.Create("keys/" + filename)
	FailOnError(err)
	defer privFile.Close()

	err = pem.Encode(privFile, privBlock)
	FailOnError(err)
}

func WriteRSAKeyToFile(filename string, key *rsa.PrivateKey) {
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	FailOnError(err)

	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}

	privFile, err := os.Create("keys/" + filename)
	FailOnError(err)
	defer privFile.Close()

	err = pem.Encode(privFile, privBlock)
	FailOnError(err)
}

func WritePublicKeyToFile(filename string, publicKey *rsa.PublicKey) []byte {
	pubBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	FailOnError(err)

	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}

	pubFile, err := os.Create("keys/" + filename)
	FailOnError(err)

	defer pubFile.Close()

	err = pem.Encode(pubFile, pubBlock)
	FailOnError(err)

	return pubBytes
}

func ReadECDSAKeyFromFile(filename string) *ecdsa.PrivateKey {
	privFile, err := os.Open("keys/" + filename)
	FailOnError(err)
	defer privFile.Close()

	info, err := privFile.Stat()
	FailOnError(err)

	size := info.Size()
	privBytes := make([]byte, size)
	_, err = privFile.Read(privBytes)
	FailOnError(err)

	privPem, _ := pem.Decode(privBytes)
	privKey, err := x509.ParseECPrivateKey(privPem.Bytes)
	FailOnError(err)

	return privKey
}

func ReadRSAKeyFromFile(filename string) *rsa.PrivateKey {
	privFile, err := os.Open("keys/" + filename)
	FailOnError(err)
	defer privFile.Close()

	info, err := privFile.Stat()
	FailOnError(err)

	size := info.Size()
	privBytes := make([]byte, size)
	_, err = privFile.Read(privBytes)
	FailOnError(err)

	privPem, _ := pem.Decode(privBytes)
	privKey, err := x509.ParsePKCS8PrivateKey(privPem.Bytes)
	FailOnError(err)

	return privKey.(*rsa.PrivateKey)
}

func ReadPublicKeyBytesFromFile(filename string) []byte {
	pubFile, err := os.Open("keys/" + filename)
	FailOnError(err)
	defer pubFile.Close()

	info, err := pubFile.Stat()
	FailOnError(err)

	size := info.Size()
	pubBytes := make([]byte, size)
	_, err = pubFile.Read(pubBytes)
	FailOnError(err)

	pubPem, _ := pem.Decode(pubBytes)

	return pubPem.Bytes
}

func ParsePublicKey(pubBytes []byte) *rsa.PublicKey {
	pubKey, err := x509.ParsePKIXPublicKey(pubBytes)
	FailOnError(err)

	return pubKey.(*rsa.PublicKey)
}

func EncryptWithRSA(data []byte, publicKey *rsa.PublicKey) []byte {
	out, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, data, nil)
	FailOnError(err)
	return out
}

func DecryptWithRSA(data []byte, privateKey *rsa.PrivateKey) []byte {
	out, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, data, nil)
	// fmt.Println(string(out))
	FailOnError(err)
	return out
}
