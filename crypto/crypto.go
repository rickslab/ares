package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"sort"
	"strings"

	"github.com/fbsobreira/gotron-sdk/pkg/common"
)

const (
	mark4 int64 = (1 << 4) - 1
)

func Sha1Hash(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum([]byte{}))
}

func Sha1Sign(strs ...string) string {
	sort.Strings(strs)
	return Sha1Hash(strings.Join(strs, ""))
}

func Sha256Hash(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum([]byte{}))
}

func Sha256Sign(strs ...string) string {
	sort.Strings(strs)
	return Sha256Hash(strings.Join(strs, ""))
}

func HS1Sign(str string, key string) string {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(str))
	return hex.EncodeToString(mac.Sum(nil))
}

func HS256Sign(str string, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(str))
	return hex.EncodeToString(mac.Sum(nil))
}

func GetEntityType(id int64) int32 {
	return int32((id >> 12) & mark4)
}

func GlcEncode(id int64) string {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(id))
	return common.EncodeCheck(b)
}

func GlcDecode(str string) (int64, error) {
	b, err := common.DecodeCheck(str)
	if err != nil {
		return 0, err
	}

	return int64(binary.LittleEndian.Uint64(b)), nil
}

func EncodeAddress(id int64, chainId int32) string {
	b := make([]byte, 20)
	binary.LittleEndian.PutUint64(b, uint64(id))
	binary.LittleEndian.PutUint64(b[8:], ^uint64(id))
	binary.LittleEndian.PutUint32(b[16:], uint32(chainId))
	return common.EncodeCheck(b)
}

func DecodeAddress(address string) (int64, int32, error) {
	b, err := common.DecodeCheck(address)
	if err != nil {
		return 0, 0, err
	}

	id := int64(binary.LittleEndian.Uint64(b))
	idr := binary.LittleEndian.Uint64(b[8:])
	if ^uint64(id) != idr {
		return 0, 0, errors.New("address invalid")
	}

	chainId := int32(binary.LittleEndian.Uint32(b[16:]))
	return id, chainId, nil
}

func EncodeAddressV2(spaceId int32, id int64, chainId int32) string {
	b := make([]byte, 24)
	binary.LittleEndian.PutUint64(b, uint64(id))
	binary.LittleEndian.PutUint64(b[8:16], ^uint64(id))
	binary.LittleEndian.PutUint32(b[16:20], uint32(spaceId))
	binary.LittleEndian.PutUint32(b[20:], uint32(chainId))
	return common.EncodeCheck(b)
}

func DecodeAddressV2(address string) (int32, int64, int32, error) {
	b, err := common.DecodeCheck(address)
	if err != nil {
		return 0, 0, 0, err
	}

	id := int64(binary.LittleEndian.Uint64(b))
	idr := binary.LittleEndian.Uint64(b[8:16])
	if ^uint64(id) != idr {
		return 0, 0, 0, errors.New("address invalid")
	}

	spaceId := int32(binary.LittleEndian.Uint32(b[16:20]))
	chainId := int32(binary.LittleEndian.Uint32(b[20:]))
	return spaceId, id, chainId, nil
}

func GetOpenId(userId int64, appId string) (string, error) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(userId))

	crypted, err := AesEncrypt(b, []byte(appId))
	if err != nil {
		return "", nil
	}
	return common.EncodeCheck(crypted), nil
}

func GetUserId(openId string, appId string) (int64, error) {
	crypted, err := common.DecodeCheck(openId)
	if err != nil {
		return 0, err
	}

	b, err := AesDecrypt(crypted, []byte(appId))
	if err != nil {
		return 0, err
	}
	return int64(binary.LittleEndian.Uint64(b)), nil
}

func pkcs7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	if length < unpadding {
		return nil
	}
	return origData[:(length - unpadding)]
}

func keyHash(key []byte) []byte {
	c := sha256.New()
	c.Write(key)
	return c.Sum(nil)
}

func AesEncrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(keyHash(key))
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	data = pkcs7Padding(data, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(data))
	blockMode.CryptBlocks(crypted, data)
	return crypted, nil
}

func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(keyHash(key))
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = pkcs7UnPadding(origData)
	return origData, nil
}

func AesCBCEncrypt(rawData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(keyHash(key))
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	rawData = pkcs7Padding(rawData, blockSize)

	cipherText := make([]byte, blockSize+len(rawData))
	iv := cipherText[:blockSize]
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[blockSize:], rawData)

	return cipherText, nil
}

func AesCBCDecrypt(encryptData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(keyHash(key))
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	if len(encryptData) < blockSize {
		return nil, errors.New("ciphertext too short")
	}

	iv := encryptData[:blockSize]
	encryptData = encryptData[blockSize:]
	if len(encryptData)%blockSize != 0 {
		return nil, errors.New("ciphertext too short")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(encryptData, encryptData)

	return pkcs7UnPadding(encryptData), nil
}
