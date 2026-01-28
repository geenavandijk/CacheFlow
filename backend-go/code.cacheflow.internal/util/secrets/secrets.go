package secrets

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/charmbracelet/log"
	"google.golang.org/api/iterator"
)

// MARK: Secret Cache
var DatabaseSecretValue string

var PublicKeyValue string
var PublicKeyVersions map[string]map[int32]string = make(map[string]map[int32]string)
var PublicKeyValueLatestVersion int32

var PrivateKeyValue string
var PrivateKeyVersions map[string]map[int32]string = make(map[string]map[int32]string)
var PrivateKeyValueLatestVersion int32

var SymmetricKeyValue string
var SymmetricKeyVersions map[string]map[int32]string = make(map[string]map[int32]string)
var SymmetricKeyValueLatestVersion int32

var EmailSecretValue string

var MassiveMainApiKeyValue string
var MassiveFuturesApiKeyValue string

var SecretsWithVersions = map[string][]int32{
}

func InitializeSecretCache() {
    // Create a new logger
    logger := log.NewWithOptions(os.Stderr, log.Options{
        ReportCaller:    true, // Report the file name and line number
        ReportTimestamp: true, // Report the timestamp
        TimeFormat:      "2006-01-02 15:04:05", // Set the time format
        Prefix:          "SECRET", // Set the prefix
    })

    var err error
    {
        DatabaseSecretValue, err = DatabaseSecret()
        if err != nil {
            logger.Fatal(err)
        }
    }
    {
        PrivateKeyValue, PrivateKeyVersions, PrivateKeyValueLatestVersion, err = PrivateKey()
        if err != nil {
            logger.Fatal(err)
        }
    }
    {
        PublicKeyValue, PublicKeyVersions, PublicKeyValueLatestVersion, err = PublicKey()
        if err != nil {
            logger.Fatal(err)
        }
    }
    {
        SymmetricKeyValue, SymmetricKeyVersions, SymmetricKeyValueLatestVersion, err = SymmetricKey()
        if err != nil {
            logger.Fatal(err)
        }
    }
    {
        EmailSecretValue, err = EmailSecret()
        if err != nil {
            logger.Fatal(err)
        }
    }
}

// MARK: Public Key
// Get the secret from Secret Manager
func PublicKey() (string, map[string]map[int32]string, int32, error) {

    // Create a new logger
    logger := log.NewWithOptions(os.Stderr, log.Options{
        ReportCaller:    true, // Report the file name and line number
        ReportTimestamp: true, // Report the timestamp
        TimeFormat:      "2006-01-02 15:04:05", // Set the time format
        Prefix:          "SECRET", // Set the prefix
    })

    ctx := context.Background()
    client, err := secretmanager.NewClient(ctx)
    if err != nil {
        logger.Fatal(err)
        return "", nil, int32(0), err
    }
    defer client.Close()

    // List all versions of the secret
    req := &secretmanagerpb.ListSecretVersionsRequest{
        Parent: "projects/cacheflow-485623/secrets/public_key",
    }

    it := client.ListSecretVersions(ctx, req)
    var versions []int32
    var latestVersion int32

    for {
        resp, err := it.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            return "", nil, int32(0), err
        }

		// if secret is in DESTROYED state, skip it
		if strings.Contains(resp.State.String(), "DESTROYED") {
			// logger.Info("Secret 'public_key' with version %d is in DESTROYED state", resp.Name)
			// logger.Info("Moving to the next version...")
			continue
		}

        versionStr := strings.Split(resp.Name, "/")
        version, err := strconv.Atoi(versionStr[len(versionStr)-1])
        if err != nil {
            return "", nil, int32(0), err
        }
        versions = append(versions, int32(version))

        if int32(version) > latestVersion {
            latestVersion = int32(version)
        }
    }

    // Loop through each version and access that, then add it to the map
    for _, version := range versions {
        reqVersion := &secretmanagerpb.AccessSecretVersionRequest{
            Name: fmt.Sprintf("projects/cacheflow-485623/secrets/public_key/versions/%d", version),
        }

        result, err := client.AccessSecretVersion(ctx, reqVersion)
        if err != nil {
			// check if secret is in DESTROYED state
			if strings.Contains(err.Error(), "DESTROYED") {
				// logger.Fatal("Secret 'public_key' with version %d is in DESTROYED state", version)
				// logger.Fatal("Moving to the next version...")
				continue
			} else {
				logger.Fatal(err)
				return "", nil, int32(0), err
			}
        }

        secret := string(result.Payload.Data)
        if _, exists := PublicKeyVersions["public_key"]; !exists {
            PublicKeyVersions["public_key"] = make(map[int32]string)
        }
        PublicKeyVersions["public_key"][version] = secret
    }

    // Access the latest version of the secret
    reqLatest := &secretmanagerpb.AccessSecretVersionRequest{
        Name: fmt.Sprintf("projects/cacheflow-485623/secrets/public_key/versions/%d", latestVersion),
    }

    result, err := client.AccessSecretVersion(ctx, reqLatest)
    if err != nil {
        logger.Fatal(err)
        return "", nil, int32(0), err
    }

//     latestVersion := int32(1)

//     tempSec := `-----BEGIN PUBLIC KEY-----
// MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA3MFfxNMLdEODn8Sjc9hd
// nAq3Irygewmaj3mucVRtetcN2nhWjvOGjqL9dBKr4myhqjGFfL0fuNKOoIU/64+Y
// eEV78vncvqkIYJ/bfJ4wtGoQnVAcGkqFmnOc8hNn/miLKoEur84D4HBf/fIb//0y
// UC76hwhF7ct/VH0KD7w1pYvnKGXBANwPtZetW8CPKTqUgux68nPWJYzv7fgqyXdG
// 0feBLqqCEk0TfywbRWmUjUspMuMAibNYotU1Wg8Vftc/i4arFi9AFdCz2HzgoNu+
// hiI0p9vcejKMTAm86I4YtklRjTbRjtQPNfOAsF0mdHtKnHY4E3WiTIs46ss6wa7V
// uwIDAQAB
// -----END PUBLIC KEY-----`

//     PublicKeyVersions["public_key"] = make(map[int32]string)
//     PublicKeyVersions["public_key"][int32(latestVersion)] = string(tempSec)

    latestSecret := string(result.Payload.Data)
    PublicKeyValue = latestSecret
    PublicKeyValueLatestVersion = int32(latestVersion)

    return latestSecret, PublicKeyVersions, latestVersion, nil
}

func GetPublicKeyWithVersion(version int32) (string, bool) {
    if secret, exists := PublicKeyVersions["public_key"][version]; exists {
        return secret, true
    }
    return "", false
}

// MARK: Private Key
// Get the secret from Secret Manager
func PrivateKey() (string, map[string]map[int32]string, int32, error) {

    logger := log.NewWithOptions(os.Stderr, log.Options{
        ReportCaller:    true, // Report the file name and line number
        ReportTimestamp: true, // Report the timestamp
        TimeFormat:      "2006-01-02 15:04:05", // Set the time format
        Prefix:          "SECRET", // Set the prefix
    })

    // Create a new client
    ctx := context.Background()
    client, err := secretmanager.NewClient(ctx)
    if err != nil {
        logger.Fatal(err)
        return "", nil, int32(0), err
    }
    defer client.Close()

    // List all versions of the secret
    req := &secretmanagerpb.ListSecretVersionsRequest{
        Parent: "projects/cacheflow-485623/secrets/private_key",
    }

    it := client.ListSecretVersions(ctx, req)
    var versions []int32
    var latestVersion int32

    for {
        resp, err := it.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            return "", nil, int32(0), err
        }

		// if secret is in DESTROYED state, skip it
		if strings.Contains(resp.State.String(), "DESTROYED") {
			// logger.Info("Secret 'private_key' with version %d is in DESTROYED state", resp.Name)
			// logger.Info("Moving to the next version...")
			continue
		}

        versionStr := strings.Split(resp.Name, "/")
        version, err := strconv.Atoi(versionStr[len(versionStr)-1])
        if err != nil {
            return "", nil, int32(0), err
        }
        versions = append(versions, int32(version))

        if int32(version) > latestVersion {
            latestVersion = int32(version)
        }
    }

    // Loop through each version and access that, then add it to the map
    for _, version := range versions {
        reqVersion := &secretmanagerpb.AccessSecretVersionRequest{
            Name: fmt.Sprintf("projects/cacheflow-485623/secrets/private_key/versions/%d", version),
        }

        result, err := client.AccessSecretVersion(ctx, reqVersion)
        if err != nil {
			// check if secret is in DESTROYED state
			if strings.Contains(err.Error(), "DESTROYED") {
				// logger.Fatal("Secret 'private_key' with version %d is in DESTROYED state", version)
				// logger.Fatal("Moving to the next version...")
				continue
			} else {
				logger.Fatal(err)
				return "", nil, int32(0), err
			}
        }

        secret := string(result.Payload.Data)
        if _, exists := PrivateKeyVersions["private_key"]; !exists {
            PrivateKeyVersions["private_key"] = make(map[int32]string)
        }
        PrivateKeyVersions["private_key"][version] = secret
    }

    // Access the latest version of the secret
    reqLatest := &secretmanagerpb.AccessSecretVersionRequest{
        Name: fmt.Sprintf("projects/cacheflow-485623/secrets/private_key/versions/%d", latestVersion),
    }

    result, err := client.AccessSecretVersion(ctx, reqLatest)
    if err != nil {
        logger.Fatal(err)
        return "", nil, int32(0), err
    }

//     tempSec := `-----BEGIN PRIVATE KEY-----
// MIIEpAIBAAKCAQEA3MFfxNMLdEODn8Sjc9hdnAq3Irygewmaj3mucVRtetcN2nhW
// jvOGjqL9dBKr4myhqjGFfL0fuNKOoIU/64+YeEV78vncvqkIYJ/bfJ4wtGoQnVAc
// GkqFmnOc8hNn/miLKoEur84D4HBf/fIb//0yUC76hwhF7ct/VH0KD7w1pYvnKGXB
// ANwPtZetW8CPKTqUgux68nPWJYzv7fgqyXdG0feBLqqCEk0TfywbRWmUjUspMuMA
// ibNYotU1Wg8Vftc/i4arFi9AFdCz2HzgoNu+hiI0p9vcejKMTAm86I4YtklRjTbR
// jtQPNfOAsF0mdHtKnHY4E3WiTIs46ss6wa7VuwIDAQABAoIBAEpqNSpYKihLCfe1
// hZPrf8K2Kf3fsAlJt3xd/FvfpfgevmDC6ArQNK+oad/S23Y8QFLq1qCx8BuRftrF
// kEKL6U3BvoKtj6gFuvk3afmcKbbneipNcuu560ehaWqA+DFedGsjmOVSgSQLPZCR
// KLcmVuuIqpMSSuBaFNYaQGYFf+AJpJGhQImul/2k2rSzqZzUdKLUHCDrQtpX4dNk
// QahYgHVVhoRmzMCdT81DE29ZpPBl8WAkKLDi0PFZDuFUCR3Y3Xr67P3ucNxIBfyO
// zYbPgjYLPi4GZ5VBdjdShY0wB1OWPjoDJniz63QffpZWXqQdblwHNCF0s19nEWne
// Kp4dg4ECgYEA/PFKDXN1xvLFdhc4XIybZApVMQsXAkPtNk6huJhSFXMj6HfUiDOR
// CHHooqvyYos9FB02hMA4X6OPDV8V2rdRckoLsmMA+Xxu5hr758qwBYNSwsh8umme
// ysLV8byd4M8fZQLZkvFJFuOB1cIasD1CrAs+QJAPzk+VKmafc3gfiZECgYEA32x7
// 8tWr+kjz9LBRhbbs80uTLOwNGwMyeOyJDMSEtteA0HPvaOOT2ftihMMHzx3JRqgg
// idHIc4yH1XJjTNWu7GfmsWa1CLAtQtFVSN3+Ahbo+2ITj7C1vORGUvw5ZtC2HdsL
// 5FL7UxP5ZnK4sYRZdDs3309hsXR5RMh/h7xq5IsCgYEA0vFrX4Jdw5ftI0nEIDLG
// 2IxhdbAY9TOu3S6AYlDti6PlW1ar8RaIKnHMn/UkBBi4zFB8igNAQEpH38/fivJ8
// wwKLdK5qPqYOWmpAkJMRbteRKo5GrrOs0M3h3GL/i19QgE37IzeES8laGulwlm/Q
// bWhG91LtS5JFBQJmkcapDnECgYEA0M/wTDW9RO9X6fG8T292p7CTfeGkIgHr8y6G
// rEMjEkv5XVltgSdpcMpM+m3Y548cJYhO2OATK0NFVI1TF2WC1foJlvW57sPRmWhS
// AAdSNhu5ZfD0/U+Xm1HyE+8dquUXwHGP/LY9pBi/+CrySR68CDCseG4VjSehi1Kk
// NnU3ph8CgYAOs4fuHf6Y3TvkshIsDASEjfEyqWzMYBtQ5J2OYIlykQgTJoP7g3XI
// OJC07JXlF/wj/MHy9NCHLyc1V4PCR6l7GWZRgfZ2jBAvEQiVaA6deDELwF1EHuTT
// wTMTxzBl3OsuVxbAZ2kmwjkgQnmBDGcdaeTK7/wZCC6j9ncIIODVcQ==
// -----END PRIVATE KEY-----`
//     PrivateKeyVersions["private_key"] = make(map[int32]string)
//     PrivateKeyVersions["private_key"][int32(1)] = string(tempSec)

    latestSecret := string(result.Payload.Data)
    PrivateKeyValue = latestSecret
    PrivateKeyValueLatestVersion = int32(latestVersion)

    return latestSecret, PrivateKeyVersions, int32(1), nil
}

func GetPrivateKeyWithVersion(version int32) (string, bool) {
    if secret, exists := PrivateKeyVersions["private_key"][version]; exists {
        return secret, true
    }
    return "", false
}

// MARK: Symmetric Key
// Get the secret from Secret Manager
func SymmetricKey() (string, map[string]map[int32]string, int32, error) {

    // Create a new logger
    logger := log.NewWithOptions(os.Stderr, log.Options{
        ReportCaller:    true, // Report the file name and line number
        ReportTimestamp: true, // Report the timestamp
        TimeFormat:      "2006-01-02 15:04:05", // Set the time format
        Prefix:          "SECRET", // Set the prefix
    })

    // Create a new client
    ctx := context.Background()
    client, err := secretmanager.NewClient(ctx)
    if err != nil {
        logger.Fatal(err)
        return "", nil, int32(0), err
    }
    defer client.Close()

    // List all versions of the secret
    req := &secretmanagerpb.ListSecretVersionsRequest{
        Parent: "projects/cacheflow-485623/secrets/symmetric_key",
    }

    it := client.ListSecretVersions(ctx, req)
    var versions []int32
    var latestVersion int32

    for {
        resp, err := it.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            return "", nil, int32(0), err
        }

		// if secret is in DESTROYED state, skip it
		if strings.Contains(resp.State.String(), "DESTROYED") {
			continue
		}

        versionStr := strings.Split(resp.Name, "/")
        version, err := strconv.Atoi(versionStr[len(versionStr)-1])
        if err != nil {
            return "", nil, int32(0), err
        }
        versions = append(versions, int32(version))

        if int32(version) > latestVersion {
            latestVersion = int32(version)
        }
    }

    // Loop through each version and access that, then add it to the map
    for _, version := range versions {
        reqVersion := &secretmanagerpb.AccessSecretVersionRequest{
            Name: fmt.Sprintf("projects/cacheflow-485623/secrets/symmetric_key/versions/%d", version),
        }

        result, err := client.AccessSecretVersion(ctx, reqVersion)
        if err != nil {
			// check if secret is in DESTROYED state
			if strings.Contains(err.Error(), "DESTROYED") {
				continue
			} else {
				logger.Fatal(err)
				return "", nil, int32(0), err
			}
        }

        secret := string(result.Payload.Data)
        if _, exists := SymmetricKeyVersions["symmetric_key"]; !exists {
            SymmetricKeyVersions["symmetric_key"] = make(map[int32]string)
        }
        SymmetricKeyVersions["symmetric_key"][version] = secret
    }

    // Access the latest version of the secret
    reqLatest := &secretmanagerpb.AccessSecretVersionRequest{
        Name: fmt.Sprintf("projects/cacheflow-485623/secrets/symmetric_key/versions/%d", latestVersion),
    }

    result, err := client.AccessSecretVersion(ctx, reqLatest)
    if err != nil {
        logger.Fatal(err)
        return "", nil, int32(0), err
    }

    latestSecret := string(result.Payload.Data)
    SymmetricKeyValue = latestSecret
    SymmetricKeyValueLatestVersion = latestVersion

    return latestSecret, SymmetricKeyVersions, latestVersion, nil
}

func GetSymmetricKeyWithVersion(version int32) (string, bool) {
	if secret, exists := SymmetricKeyVersions["symmetric_key"][version]; exists {
		return secret, true
	}
	return "", false
}

// MARK: Database Secret
// Get the secret from Secret Manager
func DatabaseSecret() (string, error) {

	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller: true, // Report the file name and line number
		ReportTimestamp: true, // Report the timestamp
		TimeFormat: "2006-01-02 15:04:05", // Set the time format
		Prefix: "SECRET", // Set the prefix
	})

	// Create a new client
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		logger.Fatal(err)
		return "", err
	}
	defer client.Close()

	// Build the request
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: "projects/cacheflow-485623/secrets/mongo_uri/versions/latest",
	}

	// Access the secret
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		logger.Fatal(err)
		return "", err
	}

	return string(result.Payload.Data), nil
}

// MARK: Email Secret
// Get the secret from Secret Manager
func EmailSecret() (string, error) {

	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller: true, // Report the file name and line number
		ReportTimestamp: true, // Report the timestamp
		TimeFormat: "2006-01-02 15:04:05", // Set the time format
		Prefix: "SECRET", // Set the prefix
	})

	// Create a new client
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		logger.Fatal(err)
		return "", err
	}
	defer client.Close()

	// Build the request
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: "projects/cacheflow-485623/secrets/email_app_key/versions/latest",
	}

	// Access the secret
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		logger.Fatal(err)
		return "", err
	}

	// Return the secret
	return string(result.Payload.Data), nil
}