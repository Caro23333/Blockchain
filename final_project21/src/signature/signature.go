package signature

import (
	"crypto/rand"
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"
	"crypto/sha256"
	"fmt"
	"log"
)

// 序列化的公钥。除了真正验证签名时，其他时候一律用这种格式储存，否则难以序列化并通过rpc传输
type JSONPublicKey struct {
	Curve string   `json:"curve"`
	X     []byte   `json:"x"`     
	Y     []byte   `json:"y"`     
}

// 将实际的公钥序列化
func PublicKeyToJSON(pubKey ecdsa.PublicKey) JSONPublicKey {
	var curveName string
	switch pubKey.Curve {
		case elliptic.P256():
			curveName = "P256"
		case elliptic.P384():
			curveName = "P384"
		case elliptic.P521():
			curveName = "P521"
		default:
			curveName = "Unknown"
	}
	return JSONPublicKey {
		Curve: curveName,
		X:     pubKey.X.Bytes(),
		Y:     pubKey.Y.Bytes(),
	}
}

// 还原出实际的公钥，用于验证签名
func JSONToPublicKey(jsonKey JSONPublicKey) (ecdsa.PublicKey, error) {
	var curve elliptic.Curve
	switch jsonKey.Curve {
		case "P256": 	
			curve = elliptic.P256()
		case "P384":
			curve = elliptic.P384()
		case "P521":
			curve = elliptic.P521()
		default:
			return ecdsa.PublicKey{}, fmt.Errorf("unsupported curve: %s", jsonKey.Curve)
	}
	res_public_key := ecdsa.PublicKey {
		Curve: curve,
		X:     new(big.Int),
		Y:     new(big.Int),
	}
	res_public_key.X.SetBytes(jsonKey.X)
	res_public_key.Y.SetBytes(jsonKey.Y)
	return res_public_key, nil
}

// 签名一个消息，返回类型是一对 []byte，它们共同组成数字签名
func SignMessage(privateKey *ecdsa.PrivateKey, message []byte) ([]byte, []byte, error) {
	hash := sha256.New()
	hash.Write(message)
	digest := hash.Sum(nil)

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, digest)
	if err != nil {
		return nil, nil, err
	}
	return r.Bytes(), s.Bytes(), nil
}

// 验证签名是否合法
func VerifySignature(publicKey JSONPublicKey, message, r, s []byte) bool {
	
	hash := sha256.New()
	hash.Write(message)
	digest := hash.Sum(nil)

	var R, S big.Int
	R.SetBytes(r)
	S.SetBytes(s)

	real_public_key, _ := JSONToPublicKey(publicKey)
	res := ecdsa.Verify(&real_public_key, digest, &R, &S)
	return res
}

// 生成一对密钥。我们使用基于椭圆曲线的ecdsa数字签名机制。ecdsa.PrivateKey中同时包含了私钥公钥的信息
func KeyGen() (*ecdsa.PrivateKey) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Printf("Error generating ECDSA key: %v", err)
		log.Fatal()
	}
	return privateKey
}