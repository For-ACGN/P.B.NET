package cert

import (
	"testing"
)

func TestGenerateTestPoolAddParallel(t *testing.T) {
	t.Run("PublicRootCACert", func(t *testing.T) {
		generateTestPoolAddCertParallel("PublicRootCA")
	})

	t.Run("PublicClientCACert", func(t *testing.T) {
		generateTestPoolAddCertParallel("PublicClientCA")
	})

	t.Run("PublicClientPair", func(t *testing.T) {
		generateTestPoolAddPairParallel("PublicClient")
	})

	t.Run("PrivateRootCAPair", func(t *testing.T) {
		generateTestPoolAddPairParallel("PrivateRootCA")
	})

	t.Run("PrivateRootCACert", func(t *testing.T) {
		generateTestPoolAddCertParallel("PrivateRootCA")
	})

	t.Run("PrivateClientCAPair", func(t *testing.T) {
		generateTestPoolAddPairParallel("PrivateClientCA")
	})

	t.Run("PrivateClientCACert", func(t *testing.T) {
		generateTestPoolAddCertParallel("PrivateClientCA")
	})

	t.Run("PrivateClientPair", func(t *testing.T) {
		generateTestPoolAddPairParallel("PrivateClient")
	})

	t.Run("all", func(t *testing.T) {
		generateTestPoolAddCertParallel("PublicRootCA")
		generateTestPoolAddCertParallel("PublicClientCA")
		generateTestPoolAddPairParallel("PublicClient")
		generateTestPoolAddPairParallel("PrivateRootCA")
		generateTestPoolAddCertParallel("PrivateRootCA")
		generateTestPoolAddPairParallel("PrivateClientCA")
		generateTestPoolAddCertParallel("PrivateClientCA")
		generateTestPoolAddPairParallel("PrivateClient")
	})
}

func TestGenerateTestPoolDeleteParallel(t *testing.T) {
	t.Run("PublicRootCACert", func(t *testing.T) {
		generateTestPoolDeleteCertParallel("PublicRootCA")
	})

	t.Run("PublicClientCACert", func(t *testing.T) {
		generateTestPoolDeleteCertParallel("PublicClientCA")
	})

	t.Run("PublicClientPair", func(t *testing.T) {
		generateTestPoolDeletePairParallel("PublicClient")
	})

	t.Run("PrivateRootCAPair", func(t *testing.T) {
		generateTestPoolDeletePairParallel("PrivateRootCA")
	})

	t.Run("PrivateClientCAPair", func(t *testing.T) {
		generateTestPoolDeletePairParallel("PrivateClientCA")
	})

	t.Run("PrivateClientPair", func(t *testing.T) {
		generateTestPoolDeletePairParallel("PrivateClient")
	})

	t.Run("all", func(t *testing.T) {
		generateTestPoolDeleteCertParallel("PublicRootCA")
		generateTestPoolDeleteCertParallel("PublicClientCA")
		generateTestPoolDeletePairParallel("PublicClient")
		generateTestPoolDeletePairParallel("PrivateRootCA")
		generateTestPoolDeletePairParallel("PrivateClientCA")
		generateTestPoolDeletePairParallel("PrivateClient")
	})
}

func TestGenerateTestPoolGetParallel(t *testing.T) {
	t.Run("PublicRootCACert", func(t *testing.T) {
		generateTestPoolGetCertsParallel("PublicRootCA")
	})

	t.Run("PublicClientCACert", func(t *testing.T) {
		generateTestPoolGetCertsParallel("PublicClientCA")
	})

	t.Run("PublicClientPair", func(t *testing.T) {
		generateTestPoolGetPairsParallel("PublicClient")
	})

	t.Run("PrivateRootCAPair", func(t *testing.T) {
		generateTestPoolGetPairsParallel("PrivateRootCA")
	})

	t.Run("PrivateRootCACert", func(t *testing.T) {
		generateTestPoolGetCertsParallel("PrivateRootCA")
	})

	t.Run("PrivateClientCAPair", func(t *testing.T) {
		generateTestPoolGetPairsParallel("PrivateClientCA")
	})

	t.Run("PrivateClientCACert", func(t *testing.T) {
		generateTestPoolGetCertsParallel("PrivateClientCA")
	})

	t.Run("PrivateClientPair", func(t *testing.T) {
		generateTestPoolGetPairsParallel("PrivateClient")
	})

	t.Run("all", func(t *testing.T) {
		generateTestPoolGetCertsParallel("PublicRootCA")
		generateTestPoolGetCertsParallel("PublicClientCA")
		generateTestPoolGetPairsParallel("PublicClient")
		generateTestPoolGetPairsParallel("PrivateRootCA")
		generateTestPoolGetCertsParallel("PrivateRootCA")
		generateTestPoolGetPairsParallel("PrivateClientCA")
		generateTestPoolGetCertsParallel("PrivateClientCA")
		generateTestPoolGetPairsParallel("PrivateClient")
	})
}

func TestGenerateTestPoolExportParallel(t *testing.T) {
	t.Run("PublicRootCACert", func(t *testing.T) {
		generateTestPoolExportCertParallel("PublicRootCA")
	})

	t.Run("PublicClientCACert", func(t *testing.T) {
		generateTestPoolExportCertParallel("PublicClientCA")
	})

	t.Run("PublicClientPair", func(t *testing.T) {
		generateTestPoolExportPairParallel("PublicClient")
	})

	t.Run("PrivateRootCAPair", func(t *testing.T) {
		generateTestPoolExportPairParallel("PrivateRootCA")
	})

	t.Run("PrivateClientCAPair", func(t *testing.T) {
		generateTestPoolExportPairParallel("PrivateClientCA")
	})

	t.Run("PrivateClientPair", func(t *testing.T) {
		generateTestPoolExportPairParallel("PrivateClient")
	})

	t.Run("all", func(t *testing.T) {
		generateTestPoolExportCertParallel("PublicRootCA")
		generateTestPoolExportCertParallel("PublicClientCA")
		generateTestPoolExportPairParallel("PublicClient")
		generateTestPoolExportPairParallel("PrivateRootCA")
		generateTestPoolExportPairParallel("PrivateClientCA")
		generateTestPoolExportPairParallel("PrivateClient")
	})
}
