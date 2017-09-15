package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"

	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"testing"

	"crypto/tls"

	test_util "github.com/cloudfoundry-incubator/credhub-cli/test"
)

const TIMESTAMP = `2016-01-01T12:00:00Z`
const UUID = `5a2edd4f-1686-4c8d-80eb-5daa866f9f86`

const VALID_ACCESS_TOKEN = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImxlZ2FjeS10b2tlbi1rZXkiLCJ0eXAiOiJKV1QifQ.eyJqdGkiOiI3NTY5MTc5OTgzOTY0M2Y4OWI2NGZlNDQ2MWU0OWJlMCIsInN1YiI6IjY3ODdiYjdlLTc4YmItNGJlNi05NTgzLTQyYTc1ZGRiYTNkNSIsInNjb3BlIjpbImNyZWRodWIud3JpdGUiLCJjcmVkaHViLnJlYWQiXSwiY2xpZW50X2lkIjoiY3JlZGh1Yl9jbGkiLCJjaWQiOiJjcmVkaHViX2NsaSIsImF6cCI6ImNyZWRodWJfY2xpIiwicmV2b2NhYmxlIjp0cnVlLCJncmFudF90eXBlIjoicGFzc3dvcmQiLCJ1c2VyX2lkIjoiNjc4N2JiN2UtNzhiYi00YmU2LTk1ODMtNDJhNzVkZGJhM2Q1Iiwib3JpZ2luIjoidWFhIiwidXNlcl9uYW1lIjoiY3JlZGh1YiIsImVtYWlsIjoiY3JlZGh1YiIsImF1dGhfdGltZSI6MTUwNDgyMTU4NSwicmV2X3NpZyI6ImU0Yjg2ODVlIiwiaWF0IjoxNTA0ODIxNTg1LCJleHAiOjE1MDQ5MDc5ODUsImlzcyI6Imh0dHBzOi8vMzQuMjA2LjIzMy4xOTU6ODQ0My9vYXV0aC90b2tlbiIsInppZCI6InVhYSIsImF1ZCI6WyJjcmVkaHViX2NsaSIsImNyZWRodWIiXX0.Ubi5k3Sy4CkcTqKvKuSkLJFpA5zfwWPlhImuwMW3HyKd6iEPuteXqnSE9r6ndvcKf_B3PS0ZduPg7v81RiZyfTGu3ObWIEdYExlmI97yfg4OQMCfo4jdr2wSzpcwixTK2FeZ2RcDklMfaSp_CTAnNcY4Lj2Jlk2eagWOCXizxsB1SHfegtGWH3FSUT5I3nJVcWAsRCMLqjHzRWYdP3CfpnMhnrjic1Ok_f2HKygiG44uUx2u3yQOV1CiZJwhxPODTuhI8X9kkQ0rLW9jW9ADVFstfXOglr-_k6tJMKMNpbXuCd_XaxOIXsxrSdFwcZw56KjuAA4iMuSfMxCbu1UyFw"
const VALID_ACCESS_TOKEN_JTI = "75691799839643f89b64fe4461e49be0"

const STRING_CREDENTIAL_OVERWRITE_REQUEST_JSON = `{"type":"%s","name":"%s","value":"%s","overwrite":%t}`
const JSON_CREDENTIAL_OVERWRITE_REQUEST_JSON = `{"type":"json","name":"%s","value":%s,"overwrite":%t}`
const CERTIFICATE_CREDENTIAL_REQUEST_JSON = `{"type":"certificate","name":"%s","value":{"ca":"%s","certificate":"%s","private_key":"%s"},"overwrite":%t}`
const CERTIFICATE_CREDENTIAL_WITH_NAMED_CA_REQUEST_JSON = `{"type":"certificate","name":"%s","value":{"ca_name":"%s","certificate":"%s","private_key":"%s"},"overwrite":%t}`
const GENERATE_CREDENTIAL_REQUEST_JSON = `{"name":"%s","type":"%s","overwrite":%t,"parameters":%s}`
const RSA_SSH_CREDENTIAL_REQUEST_JSON = `{"type":"%s","name":"%s","value":{"public_key":"%s","private_key":"%s"},"overwrite":%t}`
const GENERATE_DEFAULT_TYPE_REQUEST_JSON = `{"name":"%s","type":"password","overwrite":%t,"parameters":%s}`
const USER_GENERATE_CREDENTIAL_REQUEST_JSON = `{"name":"%s","type":"user","overwrite":%t,"parameters":%s,"value":%s}`
const USER_SET_CREDENTIAL_REQUEST_JSON = `{"type":"user","name":"%s","value":%s,"overwrite":%t}`

const JSON_CREDENTIAL_RESPONSE_JSON = `{"type":"json","id":"` + UUID + `","name":"%s","version_created_at":"` + TIMESTAMP + `","value":%s}`
const STRING_CREDENTIAL_RESPONSE_JSON = `{"type":"%s","id":"` + UUID + `","name":"%s","version_created_at":"` + TIMESTAMP + `","value":"%s"}`
const CERTIFICATE_CREDENTIAL_RESPONSE_JSON = `{"type":"certificate","id":"` + UUID + `","name":"%s","version_created_at":"` + TIMESTAMP + `","value":{"ca":"%s","certificate":"%s","private_key":"%s"}}`
const RSA_SSH_CREDENTIAL_RESPONSE_JSON = `{"type":"%s","id":"` + UUID + `","name":"%s","version_created_at":"` + TIMESTAMP + `","value":{"public_key":"%s","private_key":"%s"},"version_created_at":"` + TIMESTAMP + `"}`
const USER_CREDENTIAL_RESPONSE_JSON = `{"type":"user","id":"` + UUID + `","name":"%s","version_created_at":"` + TIMESTAMP + `","value":{"username":"%s", "password":"%s", "password_hash":"%s"}}`
const USER_WITHOUT_USERNAME_CREDENTIAL_RESPONSE_JSON = `{"type":"user","id":"` + UUID + `","name":"%s","version_created_at":"` + TIMESTAMP + `","value":{"username":null, "password":"%s", "password_hash":"%s"}}`

const STRING_CREDENTIAL_ARRAY_RESPONSE_JSON = `{"data":[` + STRING_CREDENTIAL_RESPONSE_JSON + `]}`
const JSON_CREDENTIAL_ARRAY_RESPONSE_JSON = `{"data":[` + JSON_CREDENTIAL_RESPONSE_JSON + `]}`
const CERTIFICATE_CREDENTIAL_ARRAY_RESPONSE_JSON = `{"data":[` + CERTIFICATE_CREDENTIAL_RESPONSE_JSON + `]}`
const RSA_SSH_CREDENTIAL_ARRAY_RESPONSE_JSON = `{"data":[` + RSA_SSH_CREDENTIAL_RESPONSE_JSON + `]}`
const USER_CREDENTIAL_ARRAY_RESPONSE_JSON = `{"data":[` + USER_CREDENTIAL_RESPONSE_JSON + `]}`

const STRING_CREDENTIAL_RESPONSE_YAML = "name: %s\nversion_created_at: " + TIMESTAMP + "\nid: " + UUID + "\ntype: %s\nvalue: %s"
const JSON_CREDENTIAL_RESPONSE_YAML = "id: " + UUID + "\nname: %s\ntype: json\nvalue:\n%s\nversion_created_at: " + TIMESTAMP
const CERTIFICATE_CREDENTIAL_RESPONSE_YAML = "id: " + UUID + "\nname: %s\ntype: certificate\nvalue:\n  ca: %s\n  certificate: %s\n  private_key: %s\nversion_created_at: " + TIMESTAMP
const SSH_CREDENTIAL_RESPONSE_YAML = "id: " + UUID + "\nname: %s\ntype: ssh\nvalue:\n  private_key: %s\n  public_key: %s\nversion_created_at: " + TIMESTAMP
const RSA_CREDENTIAL_RESPONSE_YAML = "id: " + UUID + "\nname: %s\ntype: rsa\nvalue:\n  private_key: %s\n  public_key: %s\nversion_created_at: " + TIMESTAMP
const USER_CREDENTIAL_RESPONSE_YAML = "id: " + UUID + "\nname: %s\ntype: user\nvalue:\n  password: %s\n  password_hash: %s\n  username: %s\nversion_created_at: " + TIMESTAMP
const USER_WITHOUT_USERNAME_CREDENTIAL_RESPONSE_YAML = "id: " + UUID + "\nname: %s\ntype: user\nvalue:\n  password: %s\n  password_hash: %s\n  username: null\nversion_created_at: " + TIMESTAMP

var responseMyValuePotatoesJson = fmt.Sprintf(STRING_CREDENTIAL_RESPONSE_JSON, "value", "my-value", "potatoes")
var responseMyPasswordPotatoesYaml = fmt.Sprintf(STRING_CREDENTIAL_RESPONSE_YAML, "my-password", "password", "potatoes")
var responseMyPasswordPotatoesJson = fmt.Sprintf(STRING_CREDENTIAL_RESPONSE_JSON, "password", "my-password", "potatoes")
var responseMyJsonFormatYaml = fmt.Sprintf(JSON_CREDENTIAL_RESPONSE_YAML, "json-secret", "  an:\n  - array\n  foo: bar\n  nested:\n    a: 1")
var responseMyJsonFormatJson = fmt.Sprintf(JSON_CREDENTIAL_RESPONSE_JSON, "json-secret", "{\"an\": [\"array\"], \"foo\": \"bar\", \"nested\": {\"a\": 1}}")
var responseMyCertificateYaml = fmt.Sprintf(CERTIFICATE_CREDENTIAL_RESPONSE_YAML, "my-secret", "my-ca", "my-cert", "my-priv")
var responseMyCertificateWithNamedCAYaml = fmt.Sprintf(CERTIFICATE_CREDENTIAL_RESPONSE_YAML, "my-secret", "known-ca-value", "my-cert", "my-priv")
var responseMyCertificateJson = fmt.Sprintf(CERTIFICATE_CREDENTIAL_RESPONSE_JSON, "my-secret", "my-ca", "my-cert", "my-priv")
var responseMyCertificateWithNewlinesJson = fmt.Sprintf(CERTIFICATE_CREDENTIAL_RESPONSE_JSON, "my-secret", `my\nca`, `my\ncert`, `my\npriv`)
var responseMySSHFooYaml = fmt.Sprintf(SSH_CREDENTIAL_RESPONSE_YAML, "foo-ssh-key", "some-private-key", "some-public-key")
var responseMySSHFooJson = fmt.Sprintf(RSA_SSH_CREDENTIAL_RESPONSE_JSON, "ssh", "foo-ssh-key", "some-public-key", "some-private-key")
var responseMySSHWithNewlinesJson = fmt.Sprintf(RSA_SSH_CREDENTIAL_RESPONSE_JSON, "ssh", "foo-ssh-key", `some\npublic\nkey`, `some\nprivate\nkey`)
var responseMyRSAFooYaml = fmt.Sprintf(RSA_CREDENTIAL_RESPONSE_YAML, "foo-rsa-key", "some-private-key", "some-public-key")
var responseMyRSAFooJson = fmt.Sprintf(RSA_SSH_CREDENTIAL_RESPONSE_JSON, "rsa", "foo-rsa-key", "some-public-key", "some-private-key")
var responseMyRSAWithNewlinesJson = fmt.Sprintf(RSA_SSH_CREDENTIAL_RESPONSE_JSON, "rsa", "foo-rsa-key", `some\npublic\nkey`, `some\nprivate\nkey`)
var responseMyUsernameYaml = fmt.Sprintf(USER_CREDENTIAL_RESPONSE_YAML, "my-username-credential", "test-password", "passw0rd-H4$h", "my-username")
var responseMyUsernameJson = fmt.Sprintf(USER_CREDENTIAL_RESPONSE_JSON, "my-username-credential", "my-username", "test-password", "passw0rd-H4$h")
var responseMySpecialCharacterJson = fmt.Sprintf(JSON_CREDENTIAL_RESPONSE_YAML, "my-character-test", "  foo: b\"ar")

func TestCommands(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Commands Suite")
}

var (
	commandPath string
	homeDir     string
	server      *Server
	authServer  *Server
)

var _ = BeforeEach(func() {
	var err error
	homeDir, err = ioutil.TempDir("", "cm-test")
	Expect(err).NotTo(HaveOccurred())

	if runtime.GOOS == "windows" {
		os.Setenv("USERPROFILE", homeDir)
	} else {
		os.Setenv("HOME", homeDir)
	}

	server = NewTlsServer("../test/server-tls-cert.pem", "../test/server-tls-key.pem")
	authServer = NewTlsServer("../test/auth-tls-cert.pem", "../test/auth-tls-key.pem")

	SetupServers(server, authServer)

	session := runCommand("api", server.URL(), "--ca-cert", "../test/server-tls-ca.pem", "--ca-cert", "../test/auth-tls-ca.pem")

	server.Reset()
	authServer.Reset()

	Eventually(session).Should(Exit(0))
})

var _ = AfterEach(func() {
	server.Close()
	authServer.Close()
	os.RemoveAll(homeDir)
})

var _ = SynchronizedBeforeSuite(func() []byte {
	executable_path, err := Build("github.com/cloudfoundry-incubator/credhub-cli", "-ldflags", "-X github.com/cloudfoundry-incubator/credhub-cli/version.Version=test-version")
	Expect(err).NotTo(HaveOccurred())
	return []byte(executable_path)
}, func(data []byte) {
	commandPath = string(data)
	test_util.CleanEnv()
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	CleanupBuildArtifacts()
})

func login() {
	authServer.AppendHandlers(
		CombineHandlers(
			VerifyRequest("POST", "/oauth/token"),
			RespondWith(http.StatusOK, `{
			"access_token":"test-access-token",
			"refresh_token":"test-refresh-token",
			"token_type":"password",
			"expires_in":123456789
			}`),
		),
	)

	runCommand("login", "-u", "test-username", "-p", "test-password")

	authServer.Reset()
}

func runCommand(args ...string) *Session {
	cmd := exec.Command(commandPath, args...)
	session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	<-session.Exited

	return session
}

func runCommandWithEnv(env []string, args ...string) *Session {
	cmd := exec.Command(commandPath, args...)
	existing := os.Environ()
	for _, env_var := range env {
		existing = append(existing, env_var)
	}
	cmd.Env = existing
	session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	<-session.Exited

	return session
}

func runCommandWithStdin(stdin io.Reader, args ...string) *Session {
	cmd := exec.Command(commandPath, args...)
	cmd.Stdin = stdin
	session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	<-session.Exited

	return session
}

func NewTlsServer(certPath, keyPath string) *Server {
	tlsServer := NewUnstartedServer()

	cert, err := ioutil.ReadFile(certPath)
	Expect(err).To(BeNil())
	key, err := ioutil.ReadFile(keyPath)
	Expect(err).To(BeNil())

	tlsCert, err := tls.X509KeyPair(cert, key)
	Expect(err).To(BeNil())

	tlsServer.HTTPTestServer.TLS = &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}

	tlsServer.HTTPTestServer.StartTLS()

	return tlsServer
}

func SetupServers(chServer, uaaServer *Server) {
	chServer.RouteToHandler("GET", "/info",
		RespondWith(http.StatusOK, `{
				"app":{"version":"my-version","name":"CredHub"},
				"auth-server":{"url":"`+uaaServer.URL()+`"}
				}`),
	)

	uaaServer.RouteToHandler("GET", "/info", RespondWith(http.StatusOK, ""))
}

func ItBehavesLikeHelp(command string, alias string, validate func(*Session)) {
	It("displays help", func() {
		session := runCommand(command, "-h")
		Eventually(session).Should(Exit(1))
		validate(session)
	})

	It("displays help using the alias", func() {
		session := runCommand(alias, "-h")
		Eventually(session).Should(Exit(1))
		validate(session)
	})
}

func ItRequiresAuthentication(args ...string) {
	It("requires authentication", func() {
		authServer.RouteToHandler("DELETE", "/oauth/token/revoke/test-refresh-token",
			RespondWith(http.StatusOK, nil),
		)

		runCommand("logout")

		session := runCommand(args...)

		Eventually(session).Should(Exit(1))
		Expect(session.Err).To(Say("You are not currently authenticated. Please log in to continue."))
	})
}

func ItAutomaticallyLogsInUsingLibrary(method string, args ...string) {
	FDescribe("automatic authentication", func() {
		AfterEach(func() {
			server.Reset()
		})

		Context("with correct environment and unauthenticated", func() {
			It("automatically authenticates", func() {
				server.RouteToHandler(method, "/api/v1/data",
					RespondWith(http.StatusOK, fmt.Sprintf(STRING_CREDENTIAL_ARRAY_RESPONSE_JSON, "value", "my-value", "potatoes")),
				)
				authServer.RouteToHandler(
					"DELETE", "/oauth/token/revoke/test-refresh-token",
					RespondWith(http.StatusOK, nil),
				)

				authServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest("POST", "/oauth/token"),
						VerifyBody([]byte(`client_id=test_client&client_secret=test_secret&grant_type=client_credentials&response_type=token`)),
						RespondWith(http.StatusOK, `{
									"access_token":"2YotnFZFEjr1zCsicMWpAA",
									"refresh_token":"erousflkajqwer",
									"token_type":"bearer",
									"expires_in":3600}`),
					),
				)

				runCommand("logout")

				session := runCommandWithEnv([]string{"CREDHUB_CLIENT=test_client", "CREDHUB_SECRET=test_secret"}, args...)

				Eventually(session).Should(Exit(0))
			})
		})

		Context("with correct environment and expired token", func() {
			It("automatically authenticates", func() {
				authServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest("POST", "/oauth/token"),
						VerifyBody([]byte(`client_id=test_client&client_secret=test_secret&grant_type=client_credentials&response_type=token`)),
						RespondWith(http.StatusOK, `{
									"access_token":"expired_token",
									"refresh_token":"erousflkajqwer",
									"token_type":"bearer",
									"expires_in":3600}`),
					),
					CombineHandlers(
						VerifyRequest("POST", "/oauth/token"),
						VerifyBody([]byte(`client_id=test_client&client_secret=test_secret&grant_type=client_credentials&response_type=token`)),
						RespondWith(http.StatusOK, `{
									"access_token":"valid_token",
									"refresh_token":"erousflkajqwer",
									"token_type":"bearer",
									"expires_in":3600}`),
					),
				)

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(method, "/api/v1/data"),
						RespondWith(http.StatusUnauthorized, `{
							"error":"access_token_expired",
							"error_description":"error description"}`),
					),
				)

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(method, "/api/v1/data"),
						RespondWith(http.StatusOK, fmt.Sprintf(STRING_CREDENTIAL_ARRAY_RESPONSE_JSON, "value", "my-value", "potatoes")),
					))

				session := runCommandWithEnv([]string{"CREDHUB_CLIENT=test_client", "CREDHUB_SECRET=test_secret"}, args...)
				Eventually(session).Should(Exit(0))
			})
		})

		Context("with expired password grant token", func() {
			It("automatically refreshes", func() {
				authServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest("POST", "/oauth/token"),
						VerifyBody([]byte(`client_id=credhub_cli&client_secret=&grant_type=refresh_token&refresh_token=test-refresh-token&response_type=token`)),
						RespondWith(http.StatusOK, `{
									"access_token":"ValidAccessToken",
									"refresh_token":"erousflkajqwer",
									"token_type":"bearer",
									"expires_in":3600}`),
					),
				)

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(method, "/api/v1/data"),
						RespondWith(http.StatusUnauthorized, `{
							"error":"access_token_expired",
							"error_description":"error description"}`),
					),
				)

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(method, "/api/v1/data"),
						RespondWith(http.StatusOK, fmt.Sprintf(STRING_CREDENTIAL_ARRAY_RESPONSE_JSON, "value", "my-value", "potatoes")),
					),
				)

				session := runCommandWithEnv([]string{}, args...)

				Eventually(session).Should(Exit(0))
			})
		})

	})
}

func ItAutomaticallyLogsIn(method string, args ...string) {
	Describe("automatic authentication", func() {
		BeforeEach(func() {
			server.RouteToHandler(method, "/api/v1/data",
				RespondWith(http.StatusOK, `{"type":"json","id":"some_uuid","name":"my-json","version_created_at":"idc","value":{"key": 1}, "credentials": [{"name": "key", "version_created_at": "something"}]}`),
			)
		})

		AfterEach(func() {
			server.Reset()
		})

		Context("with correct environment and unauthenticated", func() {
			It("automatically authenticates", func() {

				authServer.RouteToHandler(
					"DELETE", "/oauth/token/revoke/test-refresh-token",
					RespondWith(http.StatusOK, nil),
				)

				authServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest("POST", "/oauth/token"),
						VerifyBody([]byte(`client_id=test_client&client_secret=test_secret&grant_type=client_credentials&response_type=token`)),
						RespondWith(http.StatusOK, `{
								"access_token":"2YotnFZFEjr1zCsicMWpAA",
								"refresh_token":"erousflkajqwer",
								"token_type":"bearer",
								"expires_in":3600}`),
					),
				)

				runCommand("logout")

				session := runCommandWithEnv([]string{"CREDHUB_CLIENT=test_client", "CREDHUB_SECRET=test_secret"}, args...)

				Eventually(session).Should(Exit(0))
			})
		})

		Context("with correct environment and expired token", func() {
			It("automatically authenticates", func() {
				authServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest("POST", "/oauth/token"),
						VerifyBody([]byte(`client_id=test_client&client_secret=test_secret&grant_type=client_credentials&response_type=token`)),
						RespondWith(http.StatusOK, `{
								"access_token":"2YotnFZFEjr1zCsicMWpAA",
								"refresh_token":"erousflkajqwer",
								"token_type":"bearer",
								"expires_in":3600}`),
					),
				)

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(method, "/api/v1/data"),
						RespondWith(http.StatusUnauthorized, `{
						"error":"access_token_expired",
						"error_description":"error description"}`),
					),
				)

				session := runCommandWithEnv([]string{"CREDHUB_CLIENT=test_client", "CREDHUB_SECRET=test_secret"}, args...)
				Eventually(session).Should(Exit(0))
			})
		})
	})
}
