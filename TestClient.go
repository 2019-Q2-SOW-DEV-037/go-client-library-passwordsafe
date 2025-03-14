package main

import (
	"fmt"
	"os"
	"time"

	"github.com/BeyondTrust/go-client-library-passwordsafe/api/assets"
	"github.com/BeyondTrust/go-client-library-passwordsafe/api/authentication"
	"github.com/BeyondTrust/go-client-library-passwordsafe/api/databases"
	"github.com/BeyondTrust/go-client-library-passwordsafe/api/entities"
	logging "github.com/BeyondTrust/go-client-library-passwordsafe/api/logging"
	managed_accounts "github.com/BeyondTrust/go-client-library-passwordsafe/api/managed_account"
	"github.com/BeyondTrust/go-client-library-passwordsafe/api/secrets"
	"github.com/BeyondTrust/go-client-library-passwordsafe/api/utils"
	"github.com/BeyondTrust/go-client-library-passwordsafe/api/workgroups"
	backoff "github.com/cenkalti/backoff/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// main funtion
func main() {

	// create a zap logger
	//logger, _ := zap.NewProduction()
	logger, _ := zap.NewDevelopment()

	// create a zap logger wrapper
	zapLogger := logging.NewZapLogger(logger)

	// the recommended version is 3.1. If no version is specified,
	// the default API version 3.0 will be used
	apiVersion := "3.1"
	apiUrl := os.Getenv("PASSWORD_SAFE_API_URL")
	clientId := os.Getenv("PASSWORD_SAFE_CLIENT_ID")
	clientSecret := os.Getenv("PASSWORD_SAFE_CLIENT_SECRET")
	certificate := os.Getenv("PASSWORD_SAFE_CERTIFICATE")
	certificateKey := os.Getenv("PASSWORD_SAFE_CERTIFICATE_KEY")

	separator := "/"
	clientTimeOutInSeconds := 30
	verifyCa := true
	retryMaxElapsedTimeMinutes := 2
	maxFileSecretSizeBytes := 5000000

	backoffDefinition := backoff.NewExponentialBackOff()
	backoffDefinition.InitialInterval = 1 * time.Second
	backoffDefinition.MaxElapsedTime = time.Duration(retryMaxElapsedTimeMinutes) * time.Second
	backoffDefinition.RandomizationFactor = 0.5

	//certificate = os.Getenv("CERTIFICATE")
	//certificateKey = os.Getenv("CERTIFICATE_KEY")

	// Create an instance of ValidationParams
	params := utils.ValidationParams{
		ClientID:                   clientId,
		ClientSecret:               clientSecret,
		ApiUrl:                     &apiUrl,
		ApiVersion:                 apiVersion,
		ClientTimeOutInSeconds:     clientTimeOutInSeconds,
		VerifyCa:                   verifyCa,
		Logger:                     zapLogger,
		Certificate:                certificate,
		CertificateKey:             certificateKey,
		RetryMaxElapsedTimeMinutes: &retryMaxElapsedTimeMinutes,
	}

	// validate inputs
	errorsInInputs := utils.ValidateInputs(params)

	if errorsInInputs != nil {
		zapLogger.Error(fmt.Sprintf("Error: %v", errorsInInputs))
		return
	}

	// creating a http client
	httpClientObj, _ := utils.GetHttpClient(clientTimeOutInSeconds, verifyCa, certificate, certificateKey, zapLogger)

	authParams := authentication.AuthenticationParametersObj{
		HTTPClient:                 *httpClientObj,
		BackoffDefinition:          backoffDefinition,
		EndpointURL:                apiUrl,
		APIVersion:                 apiVersion,
		ClientID:                   clientId,
		ClientSecret:               clientSecret,
		ApiKey:                     "fakeone_a654+9sdf7+8we4f",
		Logger:                     zapLogger,
		RetryMaxElapsedTimeSeconds: 30,
	}

	// instantiating authenticate obj
	authenticate, _ := authentication.Authenticate(authParams)

	// authenticating
	userObject, err := authenticate.GetPasswordSafeAuthentication()
	if err != nil {
		return
	}

	// instantiating secret obj
	secretObj, _ := secrets.NewSecretObj(*authenticate, zapLogger, maxFileSecretSizeBytes)

	secretPaths := []string{"oauthgrp/credential8", "oauthgrp/file1"}

	gotSecrets, _ := secretObj.GetSecrets(secretPaths, separator)

	// WARNING: Do not log secrets in production code, the following log statement logs test secrets for testing purposes:
	zapLogger.Warn(fmt.Sprintf("%v", gotSecrets))

	// getting single secret
	gotSecret, _ := secretObj.GetSecret("oauthgrp/credential8", separator)

	// WARNING: Do not log secrets in production code, the following log statement logs test secrets for testing purposes:
	zapLogger.Warn(fmt.Sprintf("Secret Test: %v", gotSecret))

	// instantiating managed account obj
	manageAccountObj, _ := managed_accounts.NewManagedAccountObj(*authenticate, zapLogger)

	newSecretPaths := []string{"system01/managed_account01", "system01/managed_account02"}

	//managedAccountList := strings.Split(paths, ",")
	gotManagedAccounts, _ := manageAccountObj.GetSecrets(newSecretPaths, separator)

	// WARNING: Do not log secrets in production code, the following log statement logs test secrets for testing purposes:
	zapLogger.Warn(fmt.Sprintf("%v", gotManagedAccounts))

	// getting single managed account
	gotManagedAccount, _ := manageAccountObj.GetSecret("system01/managed_account01", separator)

	// WARNING: Do not log secrets in production code, the following log statement logs test secrets for testing purposes:
	zapLogger.Warn(fmt.Sprintf("%v", gotManagedAccount))

	account := entities.AccountDetails{
		AccountName:                       "ManagedAccountTest_" + uuid.New().String(),
		Password:                          "Passw0rd101!*",
		DomainName:                        "exampleDomain",
		UserPrincipalName:                 "user@example.com",
		SAMAccountName:                    "samAccount",
		DistinguishedName:                 "CN=example,CN=Users,DC=domain,DC=com",
		PrivateKey:                        "privateKey",
		Passphrase:                        "passphrase",
		PasswordFallbackFlag:              true,
		LoginAccountFlag:                  false,
		Description:                       "Sample account for testing",
		ApiEnabled:                        true,
		ReleaseNotificationEmail:          "notify@example.com",
		ChangeServicesFlag:                false,
		RestartServicesFlag:               false,
		ChangeTasksFlag:                   true,
		MaxReleaseDuration:                300000,
		ISAReleaseDuration:                180,
		MaxConcurrentRequests:             5,
		AutoManagementFlag:                false,
		DSSAutoManagementFlag:             false,
		CheckPasswordFlag:                 true,
		ResetPasswordOnMismatchFlag:       false,
		ChangePasswordAfterAnyReleaseFlag: true,
		ChangeFrequencyDays:               1,
		ChangeTime:                        "",
		NextChangeDate:                    "2023-12-01",
		UseOwnCredentials:                 true,
		ChangeWindowsAutoLogonFlag:        true,
		ChangeComPlusFlag:                 false,
		ObjectID:                          "uniqueObjectID",
	}

	// creating a managed account in system_integration_test managed system.
	createResponse, err := manageAccountObj.ManageAccountCreateFlow("system_integration_test", account)

	if err != nil {
		zapLogger.Error(fmt.Sprintf(" %v", err))
	}

	// WARNING: Do not log secrets in production code, the following log statement logs test secrets for testing purposes:
	zapLogger.Warn(fmt.Sprintf("Created Managed Account: %v", createResponse.AccountName))

	objCredential := entities.SecretCredentialDetails{
		Title:       "CREDENTIAL_" + uuid.New().String(),
		Description: "My Credential Secret Description",
		Username:    "my_user",
		Password:    "MyPass2#$!",
		OwnerType:   "User",
		Notes:       "My note",
		Owners: []entities.OwnerDetails{
			{
				GroupId: 1,
				OwnerId: userObject.UserId,
				Owner:   userObject.UserName,
				Email:   userObject.EmailAddress,
			},
		},
		Urls: []entities.UrlDetails{
			{
				Id:           uuid.New(),
				CredentialId: uuid.New(),
				Url:          "https://www.test.com/",
			},
		},
	}

	// creating a credential secret in folder1.
	createdSecret, err := secretObj.CreateSecretFlow("folder1", objCredential)

	if err != nil {
		zapLogger.Error(err.Error())
		return
	}
	// WARNING: Do not log secrets in production code, the following log statement logs test secrets for testing purposes:
	zapLogger.Debug(fmt.Sprintf("Created Credential secret: %v", createdSecret.Title))

	objText := entities.SecretTextDetails{
		Title:       "TEXT_" + uuid.New().String(),
		Description: "My Text Secret Description",
		Text:        "my_p4ssword!*2024",
		OwnerType:   "User",
		OwnerId:     userObject.UserId,
		FolderId:    uuid.New(),
		Owners: []entities.OwnerDetails{
			{
				GroupId: 1,
				OwnerId: userObject.UserId,
				Owner:   userObject.UserName,
				Email:   userObject.EmailAddress,
			},
		},
		Urls: []entities.UrlDetails{
			{
				Id:           uuid.New(),
				CredentialId: uuid.New(),
				Url:          "https://www.test.com/",
			},
		},
	}

	// creating a text secret in folder1.
	createdSecret, err = secretObj.CreateSecretFlow("folder1", objText)

	if err != nil {
		zapLogger.Error(err.Error())
		return
	}
	// WARNING: Do not log secrets in production code, the following log statement logs test secrets for testing purposes:
	zapLogger.Debug(fmt.Sprintf("Created Text secret: %v", createdSecret.Title))

	// just for test purposes, file test_secret.txt should not be present in repo.
	fileContent, err := os.ReadFile("test_secret.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	objFile := entities.SecretFileDetails{
		Title:       "FILE_" + uuid.New().String(),
		Description: "My File Secret Description",
		OwnerType:   "User",
		OwnerId:     userObject.UserId,
		Owners: []entities.OwnerDetails{
			{
				GroupId: 1,
				OwnerId: userObject.UserId,
				Owner:   userObject.UserName,
				Email:   userObject.EmailAddress,
			},
		},
		Notes:       "Notes 1",
		FileName:    "my_secret.txt",
		FileContent: string(fileContent),
		Urls: []entities.UrlDetails{
			{
				Id:           uuid.New(),
				CredentialId: uuid.New(),
				Url:          "https://www.test.com/",
			},
		},
	}

	// creating a file secret in folder1.
	createdSecret, err = secretObj.CreateSecretFlow("folder1", objFile)

	if err != nil {
		zapLogger.Error(err.Error())
		return
	}

	// WARNING: Do not log secrets in production code, the following log statement logs test secrets for testing purposes:
	zapLogger.Debug(fmt.Sprintf("Created File secret: %v", createdSecret.Title))

	folderDetails := entities.FolderDetails{
		Name:        "FOLDER_" + uuid.New().String(),
		Description: "My Folder Secret Description",
	}

	// creating a folder secret in folder1 folder.
	createdFolder, err := secretObj.CreateFolderFlow("folder1", folderDetails)

	if err != nil {
		zapLogger.Error(err.Error())
		return
	}

	// WARNING: Do not log secrets in production code, the following log statement logs test secrets for testing purposes:
	zapLogger.Debug(fmt.Sprintf("Created Folder: %v", createdFolder.Name))

	safeDetails := entities.FolderDetails{
		Name:        "SAFE_" + uuid.New().String(),
		Description: "My new Safe",
		FolderType:  "SAFE",
	}

	// creating a safe.
	createdSafe, err := secretObj.CreateFolderFlow("", safeDetails)

	if err != nil {
		zapLogger.Error(err.Error())
		return
	}

	// WARNING: Do not log secrets in production code, the following log statement logs test secrets for testing purposes:
	zapLogger.Debug(fmt.Sprintf("Created Safe: %v", createdSafe.Name))

	// instantiating workgroup obj
	workGroupObj, _ := workgroups.NewWorkGroupObj(*authenticate, zapLogger)

	workGroupDetails := entities.WorkGroupDetails{
		Name: "WORKGROUP_" + uuid.New().String(),
	}

	// creating a workgroup.
	createdWorkGroup, err := workGroupObj.CreateWorkGroupFlow(workGroupDetails)

	if err != nil {
		zapLogger.Error(err.Error())
		return
	}

	// WARNING: Do not log secrets in production code, the following log statement logs test secrets for testing purposes:
	zapLogger.Debug(fmt.Sprintf("Created Workgroup: %v", createdWorkGroup.ID))

	// instantiating asset obj
	assetObj, _ := assets.NewAssetObj(*authenticate, zapLogger)

	assetDetails := entities.AssetDetails{
		IPAddress:       "192.16.1.1",
		AssetName:       "ASSET_" + uuid.New().String(),
		DnsName:         "workstation01.local",
		DomainName:      "example.com",
		MacAddress:      "00:1A:2B:3C:4D:5E",
		AssetType:       "Laptop",
		Description:     "Asset Description",
		OperatingSystem: "Windows 11",
	}

	// creating an asset by workgroup id
	createdAsset, err := assetObj.CreateAssetByworkgroupIDFlow("6", assetDetails)

	if err != nil {
		zapLogger.Error(err.Error())
		return
	}

	// WARNING: Do not log secrets in production code, the following log statement logs test secrets for testing purposes:
	zapLogger.Debug(fmt.Sprintf("Created Asset by workgroup id: %v", createdAsset.AssetName))

	// creating an asset by workgroup name
	createdAsset, err = assetObj.CreateAssetByWorkGroupNameFlow("test", assetDetails)

	if err != nil {
		zapLogger.Error(err.Error())
		return
	}

	// WARNING: Do not log secrets in production code, the following log statement logs test secrets for testing purposes:
	zapLogger.Debug(fmt.Sprintf("Created Asset by workgroup name: %v", createdAsset.AssetName))

	// instantiating database obj
	databaseObj, _ := databases.NewDatabaseObj(*authenticate, zapLogger)

	databaseDetails := entities.DatabaseDetails{
		PlatformID:        9,
		InstanceName:      "DATABASE_" + uuid.New().String(),
		IsDefaultInstance: true,
		Port:              5432,
		Version:           "15.2",
		Template:          "StandardTemplate",
	}

	// creating a database by asset
	createdDatabase, err := databaseObj.CreateDatabaseFlow("28", databaseDetails)

	if err != nil {
		zapLogger.Error(err.Error())
		return
	}

	// WARNING: Do not log secrets in production code, the following log statement logs test secrets for testing purposes:
	zapLogger.Debug(fmt.Sprintf("Created Database by Asset: %v", createdDatabase))

	// signing out
	err = authenticate.SignOut()

	if err != nil {
		zapLogger.Error(err.Error())
		return
	}

	zapLogger.Debug(fmt.Sprintf("Signed out user: %v", userObject.UserName))

}
