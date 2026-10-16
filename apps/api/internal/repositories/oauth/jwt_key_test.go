package oauth_repo

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/suite"
	repository "github.com/unbindapp/zigzag-api/internal/repositories/repositorytest"
)

type JWTKeySuite struct {
	repository.RepositoryBaseSuite
	oauthRepo *OauthRepository
}

func (suite *JWTKeySuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.oauthRepo = NewOauthRepository(suite.DB)
}

func (suite *JWTKeySuite) TearDownTest() {
	suite.RepositoryBaseSuite.TearDownTest()
	suite.oauthRepo = nil
}

func (suite *JWTKeySuite) TestGetOrGenerateJWTPrivateKeyGenerateNew() {
	keys, err := suite.DB.JWTKey.Query().All(suite.Ctx)
	suite.NoError(err)
	suite.Len(keys, 0)

	rsaKey, pemBytes, err := suite.oauthRepo.GetOrGenerateJWTPrivateKey(suite.Ctx)
	suite.NoError(err)
	suite.NotNil(rsaKey)
	suite.NotNil(pemBytes)

	suite.Equal(2048, rsaKey.Size()*8) // 2048 bits
	suite.NotNil(rsaKey.PublicKey)

	block, _ := pem.Decode(pemBytes)
	suite.NotNil(block)
	suite.Equal("RSA PRIVATE KEY", block.Type)

	savedKeys, err := suite.DB.JWTKey.Query().All(suite.Ctx)
	suite.NoError(err)
	suite.Len(savedKeys, 1)
	suite.Equal("zigzag-default", savedKeys[0].Label)
	suite.Equal(pemBytes, savedKeys[0].PrivateKey)
}

func (suite *JWTKeySuite) TestGetOrGenerateJWTPrivateKeyUseExisting() {
	firstKey, firstPemBytes, err := suite.oauthRepo.GetOrGenerateJWTPrivateKey(suite.Ctx)
	suite.NoError(err)
	suite.NotNil(firstKey)
	suite.NotNil(firstPemBytes)

	secondKey, secondPemBytes, err := suite.oauthRepo.GetOrGenerateJWTPrivateKey(suite.Ctx)
	suite.NoError(err)
	suite.NotNil(secondKey)
	suite.NotNil(secondPemBytes)

	suite.Equal(firstKey.N, secondKey.N) // RSA modulus should be the same
	suite.Equal(firstKey.D, secondKey.D) // Private exponent should be the same
	suite.Equal(firstPemBytes, secondPemBytes)

	savedKeys, err := suite.DB.JWTKey.Query().All(suite.Ctx)
	suite.NoError(err)
	suite.Len(savedKeys, 1)
}

func (suite *JWTKeySuite) TestGetOrGenerateJWTPrivateKeyPreExisting() {
	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	suite.NoError(err)

	derStream := x509.MarshalPKCS1PrivateKey(testKey)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	testPemBytes := pem.EncodeToMemory(pemBlock)

	savedKey := suite.DB.JWTKey.Create().
		SetLabel("pre-existing-key").
		SetPrivateKey(testPemBytes).
		SaveX(suite.Ctx)

	returnedKey, returnedPemBytes, err := suite.oauthRepo.GetOrGenerateJWTPrivateKey(suite.Ctx)
	suite.NoError(err)
	suite.NotNil(returnedKey)
	suite.NotNil(returnedPemBytes)

	suite.Equal(testKey.N, returnedKey.N)
	suite.Equal(testKey.D, returnedKey.D)
	suite.Equal(testPemBytes, returnedPemBytes)
	suite.Equal(savedKey.PrivateKey, returnedPemBytes)

	keys, err := suite.DB.JWTKey.Query().All(suite.Ctx)
	suite.NoError(err)
	suite.Len(keys, 1)
	suite.Equal(savedKey.ID, keys[0].ID)
}

func (suite *JWTKeySuite) TestGetOrGenerateJWTPrivateKeyMultipleKeys() {
	testKey1, err := rsa.GenerateKey(rand.Reader, 2048)
	suite.NoError(err)
	testKey2, err := rsa.GenerateKey(rand.Reader, 2048)
	suite.NoError(err)

	derStream1 := x509.MarshalPKCS1PrivateKey(testKey1)
	pemBlock1 := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: derStream1}
	testPemBytes1 := pem.EncodeToMemory(pemBlock1)

	derStream2 := x509.MarshalPKCS1PrivateKey(testKey2)
	pemBlock2 := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: derStream2}
	testPemBytes2 := pem.EncodeToMemory(pemBlock2)

	firstSavedKey := suite.DB.JWTKey.Create().
		SetLabel("first-key").
		SetPrivateKey(testPemBytes1).
		SaveX(suite.Ctx)

	suite.DB.JWTKey.Create().
		SetLabel("second-key").
		SetPrivateKey(testPemBytes2).
		SaveX(suite.Ctx)

	returnedKey, returnedPemBytes, err := suite.oauthRepo.GetOrGenerateJWTPrivateKey(suite.Ctx)
	suite.NoError(err)
	suite.Equal(testKey1.N, returnedKey.N)
	suite.Equal(testKey1.D, returnedKey.D)
	suite.Equal(firstSavedKey.PrivateKey, returnedPemBytes)
}

func (suite *JWTKeySuite) TestGetOrGenerateJWTPrivateKeyDBClosed() {
	suite.DB.Close()
	rsaKey, pemBytes, err := suite.oauthRepo.GetOrGenerateJWTPrivateKey(suite.Ctx)
	suite.Error(err)
	suite.Nil(rsaKey)
	suite.Nil(pemBytes)
	suite.ErrorContains(err, "database is closed")
}

func (suite *JWTKeySuite) TestDecodePrivateKeySuccess() {
	originalKey, err := rsa.GenerateKey(rand.Reader, 2048)
	suite.NoError(err)

	derStream := x509.MarshalPKCS1PrivateKey(originalKey)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	pemBytes := pem.EncodeToMemory(pemBlock)

	decodedKey, returnedPemBytes, err := decodePrivateKey(pemBytes)
	suite.NoError(err)
	suite.NotNil(decodedKey)
	suite.Equal(pemBytes, returnedPemBytes)

	suite.Equal(originalKey.N, decodedKey.N)
	suite.Equal(originalKey.D, decodedKey.D)
	suite.Equal(originalKey.PublicKey.N, decodedKey.PublicKey.N)
	suite.Equal(originalKey.PublicKey.E, decodedKey.PublicKey.E)
}

func (suite *JWTKeySuite) TestDecodePrivateKeyInvalidPEM() {
	invalidPem := []byte("not a valid pem block")

	decodedKey, returnedPemBytes, err := decodePrivateKey(invalidPem)
	suite.Error(err)
	suite.Nil(decodedKey)
	suite.Nil(returnedPemBytes)
	suite.ErrorContains(err, "not a valid PEM block")
}

func (suite *JWTKeySuite) TestDecodePrivateKeyWrongPEMType() {
	pemBlock := &pem.Block{
		Type:  "CERTIFICATE", // Wrong type
		Bytes: []byte("some data"),
	}
	wrongTypePem := pem.EncodeToMemory(pemBlock)

	decodedKey, returnedPemBytes, err := decodePrivateKey(wrongTypePem)
	suite.Error(err)
	suite.Nil(decodedKey)
	suite.Nil(returnedPemBytes)
	suite.ErrorContains(err, "not a valid PEM block")
}

func (suite *JWTKeySuite) TestDecodePrivateKeyInvalidRSAData() {
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: []byte("invalid rsa key data"),
	}
	invalidRsaPem := pem.EncodeToMemory(pemBlock)

	decodedKey, returnedPemBytes, err := decodePrivateKey(invalidRsaPem)
	suite.Error(err)
	suite.Nil(decodedKey)
	suite.Nil(returnedPemBytes)
	suite.ErrorContains(err, "failed to parse RSA private key")
}

func (suite *JWTKeySuite) TestDecodePrivateKeyEmptyInput() {
	decodedKey, returnedPemBytes, err := decodePrivateKey([]byte{})
	suite.Error(err)
	suite.Nil(decodedKey)
	suite.Nil(returnedPemBytes)
	suite.ErrorContains(err, "not a valid PEM block")
}

func (suite *JWTKeySuite) TestDecodePrivateKeyNilInput() {
	decodedKey, returnedPemBytes, err := decodePrivateKey(nil)
	suite.Error(err)
	suite.Nil(decodedKey)
	suite.Nil(returnedPemBytes)
	suite.ErrorContains(err, "not a valid PEM block")
}

func TestJWTKeySuite(t *testing.T) {
	suite.Run(t, new(JWTKeySuite))
}
