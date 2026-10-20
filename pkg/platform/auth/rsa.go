package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
)

const (
	rsaPrivateKeyObject = "login-rsa-private.pem"
	rsaPublicKeyObject  = "login-rsa-public.pem"
)

type rsaManager struct {
	privateKey *rsa.PrivateKey
	publicB64  string
}

// newRSAManager loads the RSA key pair from MinIO, generating it only when the
// objects do not exist yet. It fails (and thus blocks startup) when MinIO is
// unreachable.
func newRSAManager(ctx context.Context, mc *minio.Client, bucket string) (*rsaManager, error) {
	if mc == nil {
		return nil, fmt.Errorf("minio client is not configured for RSA key storage")
	}
	if bucket == "" {
		bucket = "harns"
	}

	exists, err := mc.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("check minio bucket %q failed: %w", bucket, err)
	}
	if !exists {
		if err := mc.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create minio bucket %q failed: %w", bucket, err)
		}
	}

	privExists, err := objectExists(ctx, mc, bucket, rsaPrivateKeyObject)
	if err != nil {
		return nil, err
	}
	pubExists, err := objectExists(ctx, mc, bucket, rsaPublicKeyObject)
	if err != nil {
		return nil, err
	}

	if !privExists || !pubExists {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("generate rsa key failed: %w", err)
		}
		privPEM, err := marshalPrivateKeyPEM(key)
		if err != nil {
			return nil, err
		}
		pubPEM, err := marshalPublicKeyPEM(&key.PublicKey)
		if err != nil {
			return nil, err
		}
		if !privExists {
			if err := uploadObject(ctx, mc, bucket, rsaPrivateKeyObject, privPEM); err != nil {
				return nil, err
			}
		}
		if !pubExists {
			if err := uploadObject(ctx, mc, bucket, rsaPublicKeyObject, pubPEM); err != nil {
				return nil, err
			}
		}
	}

	privPEM, err := downloadObject(ctx, mc, bucket, rsaPrivateKeyObject)
	if err != nil {
		return nil, err
	}
	pubPEM, err := downloadObject(ctx, mc, bucket, rsaPublicKeyObject)
	if err != nil {
		return nil, err
	}

	priv, err := parsePrivateKey(privPEM)
	if err != nil {
		return nil, err
	}
	pub, err := parsePublicKey(pubPEM)
	if err != nil {
		return nil, err
	}
	pubDER, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, fmt.Errorf("marshal public key failed: %w", err)
	}

	return &rsaManager{
		privateKey: priv,
		publicB64:  base64.StdEncoding.EncodeToString(pubDER),
	}, nil
}

func objectExists(ctx context.Context, mc *minio.Client, bucket, object string) (bool, error) {
	_, err := mc.StatObject(ctx, bucket, object, minio.StatObjectOptions{})
	if err == nil {
		return true, nil
	}
	resp := minio.ToErrorResponse(err)
	if resp.Code == "NoSuchKey" || resp.StatusCode == 404 {
		return false, nil
	}
	return false, fmt.Errorf("stat minio object %q failed: %w", object, err)
}

func uploadObject(ctx context.Context, mc *minio.Client, bucket, object string, data []byte) error {
	_, err := mc.PutObject(ctx, bucket, object, bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{ContentType: "application/x-pem-file"})
	if err != nil {
		return fmt.Errorf("upload rsa key %q failed: %w", object, err)
	}
	return nil
}

func downloadObject(ctx context.Context, mc *minio.Client, bucket, object string) ([]byte, error) {
	obj, err := mc.GetObject(ctx, bucket, object, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("download rsa key %q failed: %w", object, err)
	}
	defer obj.Close()
	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("read rsa key %q failed: %w", object, err)
	}
	return data, nil
}

func marshalPrivateKeyPEM(key *rsa.PrivateKey) ([]byte, error) {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("marshal private key failed: %w", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), nil
}

func marshalPublicKeyPEM(pub *rsa.PublicKey) ([]byte, error) {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, fmt.Errorf("marshal public key failed: %w", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}), nil
}

func parsePrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("invalid rsa private key pem")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse rsa private key failed: %w", err)
	}
	priv, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an rsa private key")
	}
	return priv, nil
}

func parsePublicKey(data []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("invalid rsa public key pem")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse rsa public key failed: %w", err)
	}
	pub, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an rsa public key")
	}
	return pub, nil
}

// PublicKeyBase64 returns the X.509 DER public key encoded in base64.
func (m *rsaManager) PublicKeyBase64() string {
	return m.publicB64
}

// Decrypt decrypts a base64 encoded RSA-OAEP(SHA-256) ciphertext.
func (m *rsaManager) Decrypt(ciphertextB64 string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", err
	}
	plain, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, m.privateKey, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
