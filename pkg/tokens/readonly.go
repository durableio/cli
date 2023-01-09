package tokens

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/durableio/cli/gen/proto/token/v1"
	"google.golang.org/protobuf/proto"
)

type TokenManager struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

func Bootstrap() (*TokenManager, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	tm := &TokenManager{
		privateKey,
		publicKey,
	}

	return tm, nil
}

func (tm *TokenManager) CreateWorkflowToken(workflowId string) (string, error) {
	token, err := proto.Marshal(&tokenv1.ReadWorkflowToken{
		Version:    "v1",
		WorkflowId: workflowId,
		ExpireAt:   time.Now().Add(24 * time.Hour).Unix(),
	})
	if err != nil {
		return "", err
	}

	signature := ed25519.Sign(tm.privateKey, token)

	signedToken, err := proto.Marshal(&tokenv1.SignedToken{
		Token:     token,
		Signature: signature,
	})
	if err != nil {
		return "", err
	}

	encoded := base58.Encode(signedToken)
	return fmt.Sprintf("wf_r_%s", encoded), nil

}

// ParseWorkflowToken verifies the signature and returns the workflowId
func (tm *TokenManager) ParseWorkflowToken(token string) (string, error) {
	token = strings.TrimPrefix(token, "wf_r_")
	token = string(base58.Decode(token))

	signedToken := &tokenv1.SignedToken{}
	err := proto.Unmarshal([]byte(token), signedToken)
	if err != nil {
		return "", err
	}
	isValid := ed25519.Verify(tm.publicKey, signedToken.Token, signedToken.Signature)
	if !isValid {
		return "", fmt.Errorf("Signature not valid")
	}

	readWorkflowToken := &tokenv1.ReadWorkflowToken{}
	err = proto.Unmarshal(signedToken.Token, readWorkflowToken)
	if err != nil {
		return "", err
	}

	if time.Now().After(time.Unix(readWorkflowToken.ExpireAt, 0)){
		return "", fmt.Errorf("This token has expired")
	}
	return readWorkflowToken.WorkflowId, nil

}
