// Copyright (c) 2018 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotexproject/iotex-core/pkg/keypair"
	"github.com/iotexproject/iotex-core/test/testaddress"
)

func TestNewDefaultConfig(t *testing.T) {
	// Default config doesn't have block producer addr setup
	cfg, err := New()
	require.NotNil(t, err)
	require.Nil(t, cfg)
	require.Equal(t, ErrInvalidCfg, errors.Cause(err))
}

func TestNewConfigWithoutValidation(t *testing.T) {
	cfg, err := New(DoNotValidate)
	require.Nil(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, Default, *cfg)
}

func TestNewConfigWithWrongConfigPath(t *testing.T) {
	_overwritePath = "wrong_path"
	defer func() { _overwritePath = "" }()

	cfg, err := New()
	require.NotNil(t, err)
	require.Nil(t, cfg)
	require.Contains(t, err.Error(), "open wrong_path: no such file or directory")
}

func TestNewConfigWithOverride(t *testing.T) {
	cfgStr := fmt.Sprintf(`
nodeType: %s
chain:
    producerPrivKey: "%s"
    producerPubKey: "%s"
`,
		DelegateType,
		keypair.EncodePrivateKey(testaddress.Addrinfo["alfa"].PrivateKey),
		keypair.EncodePublicKey(testaddress.Addrinfo["alfa"].PublicKey),
	)
	_overwritePath = filepath.Join(os.TempDir(), "config.yaml")
	err := ioutil.WriteFile(_overwritePath, []byte(cfgStr), 0666)
	require.NoError(t, err)
	defer func() {
		err := os.Remove(_overwritePath)
		_overwritePath = ""
		require.Nil(t, err)
	}()

	cfg, err := New()
	require.Nil(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, DelegateType, cfg.NodeType)
	require.Equal(t, keypair.EncodePrivateKey(testaddress.Addrinfo["alfa"].PrivateKey), cfg.Chain.ProducerPrivKey)
	require.Equal(t, keypair.EncodePublicKey(testaddress.Addrinfo["alfa"].PublicKey), cfg.Chain.ProducerPubKey)
}

func TestNewConfigWithSecret(t *testing.T) {
	cfgStr := fmt.Sprintf(`
nodeType: %s
chain:
    producerPrivKey: "%s"
    producerPubKey: "%s"
`,
		DelegateType,
		keypair.EncodePrivateKey(testaddress.Addrinfo["alfa"].PrivateKey),
		keypair.EncodePublicKey(testaddress.Addrinfo["alfa"].PublicKey),
	)
	_overwritePath = filepath.Join(os.TempDir(), "config.yaml")
	err := ioutil.WriteFile(_overwritePath, []byte(cfgStr), 0666)
	require.NoError(t, err)
	defer func() {
	}()

	cfgStr = fmt.Sprintf(`
chain:
    producerPrivKey: "%s"
    producerPubKey: "%s"
`,
		keypair.EncodePrivateKey(testaddress.Addrinfo["echo"].PrivateKey),
		keypair.EncodePublicKey(testaddress.Addrinfo["echo"].PublicKey),
	)
	_secretPath = filepath.Join(os.TempDir(), "secret.yaml")
	err = ioutil.WriteFile(_secretPath, []byte(cfgStr), 0666)
	require.NoError(t, err)

	defer func() {
		err := os.Remove(_overwritePath)
		require.Nil(t, err)
		_overwritePath = ""
		err = os.Remove(_secretPath)
		require.Nil(t, err)
		_secretPath = ""
	}()

	cfg, err := New()
	require.Nil(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, DelegateType, cfg.NodeType)
	require.Equal(t, keypair.EncodePrivateKey(testaddress.Addrinfo["echo"].PrivateKey), cfg.Chain.ProducerPrivKey)
	require.Equal(t, keypair.EncodePublicKey(testaddress.Addrinfo["echo"].PublicKey), cfg.Chain.ProducerPubKey)
}

func TestNewConfigWithLookupEnv(t *testing.T) {
	oldEnv, oldExist := os.LookupEnv("IOTEX_TEST_NODE_TYPE")
	err := os.Setenv("IOTEX_TEST_NODE_TYPE", DelegateType)
	require.Nil(t, err)

	cfgStr := fmt.Sprintf(`
nodeType: ${IOTEX_TEST_NODE_TYPE:"lightweight"}
chain:
    producerPrivKey: "%s"
    producerPubKey: "%s"
`,
		keypair.EncodePrivateKey(testaddress.Addrinfo["alfa"].PrivateKey),
		keypair.EncodePublicKey(testaddress.Addrinfo["alfa"].PublicKey),
	)
	_overwritePath = filepath.Join(os.TempDir(), "config.yaml")
	err = ioutil.WriteFile(_overwritePath, []byte(cfgStr), 0666)
	require.NoError(t, err)

	defer func() {
		err := os.Remove(_overwritePath)
		require.Nil(t, err)
		_overwritePath = ""
		if oldExist {
			err = os.Setenv("IOTEX_TEST_NODE_TYPE", oldEnv)
		} else {
			err = os.Unsetenv("IOTEX_TEST_NODE_TYPE")
		}
		require.Nil(t, err)
	}()

	cfg, err := New()
	require.Nil(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, DelegateType, cfg.NodeType)

	err = os.Unsetenv("IOTEX_TEST_NODE_TYPE")
	require.Nil(t, err)

	cfg, err = New()
	require.Nil(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, LightweightType, cfg.NodeType)
}

func TestValidateKeyPair(t *testing.T) {
	cfg := Default
	cfg.Chain.ProducerPubKey = "hello world"
	cfg.Chain.ProducerPrivKey = "world hello"
	err := ValidateKeyPair(&cfg)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "encoding/hex:"), err.Error())

	cfg.Chain.ProducerPubKey = keypair.EncodePublicKey(testaddress.Addrinfo["alfa"].PublicKey)
	cfg.Chain.ProducerPrivKey = keypair.EncodePrivateKey(testaddress.Addrinfo["bravo"].PrivateKey)
	err = ValidateKeyPair(&cfg)
	assert.NotNil(t, err)
	require.Equal(t, ErrInvalidCfg, errors.Cause(err))
	require.True(
		t,
		strings.Contains(err.Error(), "block producer has unmatched pubkey and prikey"),
	)
}

func TestValidateExplorer(t *testing.T) {
	cfg := Default
	cfg.Explorer.Enabled = true
	cfg.Explorer.TpsWindow = 0
	err := ValidateExplorer(&cfg)
	require.NotNil(t, err)
	require.Equal(t, ErrInvalidCfg, errors.Cause(err))
	require.True(
		t,
		strings.Contains(err.Error(), "tps window is not a positive integer when the explorer is enabled"),
	)
}

func TestValidateChain(t *testing.T) {
	cfg := Default
	cfg.Chain.NumCandidates = 0

	err := ValidateChain(&cfg)
	require.Error(t, err)
	require.Equal(t, ErrInvalidCfg, errors.Cause(err))
	require.True(
		t,
		strings.Contains(err.Error(), "candidate number should be greater than 0"),
	)

	cfg.NodeType = DelegateType
	cfg.Consensus.Scheme = RollDPoSScheme
	cfg.Consensus.RollDPoS.NumDelegates = 5
	cfg.Chain.NumCandidates = 3
	err = ValidateChain(&cfg)
	require.Error(t, err)
	require.Equal(t, ErrInvalidCfg, errors.Cause(err))
	require.True(
		t,
		strings.Contains(err.Error(), "candidate number should be greater than or equal to delegate number"),
	)
}

func TestValidateConsensusScheme(t *testing.T) {
	cfg := Default
	cfg.NodeType = FullNodeType
	cfg.Consensus.Scheme = RollDPoSScheme
	err := ValidateConsensusScheme(&cfg)
	require.NotNil(t, err)
	require.Equal(t, ErrInvalidCfg, errors.Cause(err))
	require.True(
		t,
		strings.Contains(err.Error(), "consensus scheme of fullnode should be NOOP"),
	)

	cfg.NodeType = LightweightType
	err = ValidateConsensusScheme(&cfg)
	assert.NotNil(t, err)
	require.Equal(t, ErrInvalidCfg, errors.Cause(err))
	require.True(
		t,
		strings.Contains(err.Error(), "consensus scheme of lightweight node should be NOOP"),
	)

	cfg.NodeType = "Unknown"
	err = ValidateConsensusScheme(&cfg)
	require.NotNil(t, err)
	require.Equal(t, ErrInvalidCfg, errors.Cause(err))
	require.True(
		t,
		strings.Contains(err.Error(), "unknown node type"),
	)
}

func TestValidateDispatcher(t *testing.T) {
	cfg := Default
	cfg.Dispatcher.EventChanSize = 0
	err := ValidateDispatcher(&cfg)
	require.NotNil(t, err)
	require.Equal(t, ErrInvalidCfg, errors.Cause(err))
	require.True(
		t,
		strings.Contains(err.Error(), "dispatcher event chan size should be greater than 0"),
	)
}

func TestValidateRollDPoS(t *testing.T) {
	cfg := Default
	cfg.NodeType = DelegateType
	cfg.Consensus.Scheme = RollDPoSScheme
	cfg.Consensus.RollDPoS.EventChanSize = 0
	err := ValidateRollDPoS(&cfg)
	require.NotNil(t, err)
	require.Equal(t, ErrInvalidCfg, errors.Cause(err))
	require.True(
		t,
		strings.Contains(err.Error(), "roll-DPoS event chan size should be greater than 0"),
	)

	cfg.Consensus.RollDPoS.EventChanSize = 1
	cfg.Consensus.RollDPoS.NumDelegates = 0
	err = ValidateRollDPoS(&cfg)
	require.NotNil(t, err)
	require.Equal(t, ErrInvalidCfg, errors.Cause(err))
	require.True(
		t,
		strings.Contains(err.Error(), "roll-DPoS event delegate number should be greater than 0"),
	)

	cfg.Consensus.RollDPoS.NumDelegates = 1
	cfg.Consensus.RollDPoS.EnableDummyBlock = true
	cfg.Consensus.RollDPoS.TimeBasedRotation = true
	err = ValidateRollDPoS(&cfg)
	require.NotNil(t, err)
	require.Equal(t, ErrInvalidCfg, errors.Cause(err))
	require.True(
		t,
		strings.Contains(err.Error(), "roll-DPoS should enable dummy block when doing time based rotation"),
	)
}

func TestValidateNetwork(t *testing.T) {
	cfg := Default
	cfg.Network.PeerDiscovery = false
	err := ValidateNetwork(&cfg)
	require.Error(t, err)
	require.Equal(t, ErrInvalidCfg, errors.Cause(err))
	require.True(
		t,
		strings.Contains(err.Error(), "either peer discover should be enabled or a topology should be given"),
	)
}

func TestValidateActPool(t *testing.T) {
	cfg := Default
	cfg.ActPool.MaxNumActsPerAcct = 0
	err := ValidateActPool(&cfg)
	require.NotNil(t, err)
	require.Equal(t, ErrInvalidCfg, errors.Cause(err))
	require.True(
		t,
		strings.Contains(
			err.Error(),
			"maximum number of actions per pool or per account cannot be zero or negative",
		),
	)

	cfg.ActPool.MaxNumActsPerAcct = 100
	cfg.ActPool.MaxNumActsPerPool = 0
	err = ValidateActPool(&cfg)
	require.NotNil(t, err)
	require.Equal(t, ErrInvalidCfg, errors.Cause(err))
	require.True(
		t,
		strings.Contains(
			err.Error(),
			"maximum number of actions per pool or per account cannot be zero or negative",
		),
	)

	cfg.ActPool.MaxNumActsPerPool = 99
	err = ValidateActPool(&cfg)
	require.NotNil(t, err)
	require.Equal(t, ErrInvalidCfg, errors.Cause(err))
	require.True(
		t,
		strings.Contains(
			err.Error(),
			"maximum number of actions per pool cannot be less than maximum number of actions per account",
		),
	)
}

func TestCheckNodeType(t *testing.T) {
	cfg := Default
	require.True(t, cfg.IsFullnode())
	require.False(t, cfg.IsDelegate())
	require.False(t, cfg.IsLightweight())

	cfg.NodeType = DelegateType
	require.False(t, cfg.IsFullnode())
	require.True(t, cfg.IsDelegate())
	require.False(t, cfg.IsLightweight())

	cfg.NodeType = LightweightType
	require.False(t, cfg.IsFullnode())
	require.False(t, cfg.IsDelegate())
	require.True(t, cfg.IsLightweight())
}
