// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handler

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	"yunion.io/x/pkg/errors"

	"yunion.io/x/onecloud/pkg/apigateway/clientman"
	"yunion.io/x/onecloud/pkg/apigateway/constants"
	"yunion.io/x/onecloud/pkg/apigateway/options"
	policytool "yunion.io/x/onecloud/pkg/apigateway/policy"
	"yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/appctx"
	"yunion.io/x/onecloud/pkg/appsrv"
	"yunion.io/x/onecloud/pkg/cloudcommon/policy"
	"yunion.io/x/onecloud/pkg/httperrors"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/auth"
	"yunion.io/x/onecloud/pkg/mcclient/modulebase"
	"yunion.io/x/onecloud/pkg/mcclient/modules"
	"yunion.io/x/onecloud/pkg/util/httputils"
	"yunion.io/x/onecloud/pkg/util/netutils2"
	"yunion.io/x/onecloud/pkg/util/rbacutils"
	"yunion.io/x/onecloud/pkg/util/stringutils2"
)

func AppContextToken(ctx context.Context) mcclient.TokenCredential {
	val := ctx.Value(appctx.AppContextKey(constants.AUTH_TOKEN))
	if val == nil {
		return nil
	}
	return val.(mcclient.TokenCredential)
}

type AuthHandlers struct {
	*SHandlers
	preLoginHook PreLoginFunc
}

func NewAuthHandlers(prefix string, preLoginHook PreLoginFunc) *AuthHandlers {
	return &AuthHandlers{
		SHandlers:    NewHandlers(prefix),
		preLoginHook: preLoginHook,
	}
}

func (h *AuthHandlers) AddMethods() {
	// no middleware handler
	h.AddByMethod(GET, nil,
		NewHP(h.getRegions, "regions"),
		NewHP(h.getIdpSsoRedirectUri, "sso", "redirect", "<idp_id>"),
		NewHP(h.listTotpRecoveryQuestions, "recovery"),
		NewHP(h.handleSsoLogin, "ssologin"),
		NewHP(handleOIDCAuth, "oidc", "auth"),
		NewHP(handleOIDCConfiguration, "oidc", ".well-known", "openid-configuration"),
		NewHP(handleOIDCJWKeys, "oidc", "keys"),
	)
	h.AddByMethod(POST, nil,
		NewHP(h.initTotpSecrets, "initcredential"),
		NewHP(h.resetTotpSecrets, "credential"),
		NewHP(h.validatePasscode, "passcode"),
		NewHP(h.resetTotpRecoveryQuestions, "recovery"),
		NewHP(h.postLoginHandler, "login"),
		NewHP(h.postLogoutHandler, "logout"),
		NewHP(h.handleSsoLogin, "ssologin"),
		NewHP(handleOIDCToken, "oidc", "token"),
	)

	// auth middleware handler
	h.AddByMethod(GET, FetchAuthToken,
		NewHP(h.getUser, "user"),
		NewHP(h.getPermissionDetails, "permissions"),
		NewHP(h.getAdminResources, "admin_resources"),
		NewHP(h.getResources, "scoped_resources"),
		NewHP(fetchIdpBasicConfig, "idp", "<idp_id>", "info"),
		NewHP(fetchIdpSAMLMetadata, "idp", "<idp_id>", "saml-metadata"),
		NewHP(handleOIDCUserInfo, "oidc", "user"),
	)
	h.AddByMethod(POST, FetchAuthToken,
		NewHP(h.resetUserPassword, "password"),
		NewHP(h.getPermissionDetails, "permissions"),
		NewHP(h.doCreatePolicies, "policies"),
		NewHP(handleUnlinkIdp, "unlink-idp"),
	)
	h.AddByMethod(PATCH, FetchAuthToken,
		NewHP(h.doPatchPolicy, "policies", "<policy_id>"),
	)
	h.AddByMethod(DELETE, FetchAuthToken,
		NewHP(h.doDeletePolicies, "policies"),
	)
}

func (h *AuthHandlers) Bind(app *appsrv.Application) {
	h.AddMethods()
	h.SHandlers.Bind(app)
}

func (h *AuthHandlers) GetRegionsResponse(ctx context.Context, w http.ResponseWriter, req *http.Request) (*jsonutils.JSONDict, error) {
	var currentDomain string
	var createUser bool
	qs, _ := jsonutils.ParseQueryString(req.URL.RawQuery)
	if qs != nil {
		currentDomain, _ = qs.GetString("domain")
		createUser = jsonutils.QueryBoolean(qs, "auto_create_user", true)
	}

	adminToken := auth.AdminCredential()
	if adminToken == nil {
		return nil, errors.Error("failed to get admin credential")
	}
	regions := adminToken.GetRegions()
	if len(regions) == 0 {
		return nil, errors.Error("region is empty")
	}
	regionsJson := jsonutils.NewStringArray(regions)
	s := auth.GetAdminSession(ctx, regions[0], "")
	filters := jsonutils.NewDict()
	if len(currentDomain) > 0 {
		filters.Add(jsonutils.NewString(currentDomain), "id")
	}
	filters.Add(jsonutils.NewInt(1000), "limit")
	result, e := modules.Domains.List(s, filters)
	if e != nil {
		return nil, errors.Wrap(e, "list domain")
	}
	domains := jsonutils.NewArray()
	for _, d := range result.Data {
		dn, e := d.Get("name")
		if e == nil {
			if status, err := d.Bool("enabled"); err == nil && status {
				domains.Add(dn)
			}
		}
	}
	resp := jsonutils.NewDict()
	resp.Add(domains, "domains")
	resp.Add(regionsJson, "regions")

	filters = jsonutils.NewDict()
	filters.Add(jsonutils.JSONTrue, "enabled")
	if len(currentDomain) == 0 {
		currentDomain = "all"
	}
	filters.Add(jsonutils.NewString(currentDomain), "sso_domain")
	filters.Add(jsonutils.NewString("system"), "scope")
	filters.Add(jsonutils.NewInt(1000), "limit")
	if !createUser {
		filters.Add(jsonutils.JSONFalse, "auto_create_user")
	}
	idps, err := modules.IdentityProviders.List(s, filters)
	if err != nil {
		return nil, errors.Wrap(err, "list idp")
	}
	retIdps := make([]jsonutils.JSONObject, 0)
	for i := range idps.Data {
		retIdp := idps.Data[i].(*jsonutils.JSONDict).CopyIncludes("id", "name", "driver", "template", "icon_uri")
		retIdps = append(retIdps, retIdp)
	}

	resp.Add(jsonutils.NewArray(retIdps...), "idps")

	return resp, nil
}

func (h *AuthHandlers) getRegions(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	resp, err := h.GetRegionsResponse(ctx, w, req)
	if err != nil {
		httperrors.GeneralServerError(ctx, w, err)
		return
	}
	appsrv.SendJSON(w, resp)
}

func (h *AuthHandlers) getUser(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	data, err := getUserInfo(ctx, req)
	if err != nil {
		httperrors.NotFoundError(ctx, w, err.Error())
		return
	}
	body := jsonutils.NewDict()
	body.Add(data, "data")

	appsrv.SendJSON(w, body)
}

func (h *AuthHandlers) initTotpSecrets(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	initTotpSecrets(ctx, w, req)
}

func (h *AuthHandlers) resetTotpSecrets(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	resetTotpSecrets(ctx, w, req)
}

func (h *AuthHandlers) validatePasscode(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	validatePasscodeHandler(ctx, w, req)
}

func (h *AuthHandlers) resetTotpRecoveryQuestions(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	resetTotpRecoveryQuestions(ctx, w, req)
}

func (h *AuthHandlers) listTotpRecoveryQuestions(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	listTotpRecoveryQuestions(ctx, w, req)
}

// 返回 token及totp验证状态
func doTenantLogin(ctx context.Context, req *http.Request, body jsonutils.JSONObject) (mcclient.TokenCredential, *clientman.SAuthToken, error) {
	tenantId, err := body.GetString("tenantId")
	if err != nil {
		return nil, nil, httperrors.NewInputParameterError("not found tenantId in body")
	}
	token, authToken, err := fetchAuthInfo(ctx, req)
	if err != nil {
		return nil, nil, errors.Wrapf(httperrors.ErrInvalidCredential, "fetchAuthToken fail %s", err)
	}
	if !authToken.IsTotpVerified() {
		return nil, nil, errors.Wrap(httperrors.ErrInvalidCredential, "TOTP authentication failed")
	}

	ntoken, err := auth.Client().SetProject(tenantId, "", "", token)
	if err != nil {
		return nil, nil, httperrors.NewInvalidCredentialError("failed to change project")
	}

	authToken.SetToken(ntoken.GetTokenString())
	return ntoken, authToken, nil
}

func fetchUserInfoFromToken(ctx context.Context, req *http.Request, token mcclient.TokenCredential) (jsonutils.JSONObject, error) {
	s := auth.GetAdminSession(ctx, FetchRegion(req), "")
	info, err := modules.UsersV3.Get(s, token.GetUserId(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "UsersV3.Get")
	}
	return info, nil
}

func isUserEnableTotp(userInfo jsonutils.JSONObject) bool {
	return jsonutils.QueryBoolean(userInfo, "enable_mfa", false)
}

func (h *AuthHandlers) doCredentialLogin(ctx context.Context, req *http.Request, body jsonutils.JSONObject) (mcclient.TokenCredential, error) {
	var token mcclient.TokenCredential
	var err error
	var tenant string
	// log.Debugf("doCredentialLogin body: %s", body)
	cliIp := netutils2.GetHttpRequestIp(req)
	if body.Contains("username") {
		uname, _ := body.GetString("username")

		if h.preLoginHook != nil {
			if err = h.preLoginHook(ctx, req, uname, body); err != nil {
				return nil, err
			}
		}

		var passwd string
		passwd, err = body.GetString("password")
		if err != nil {
			return nil, httperrors.NewInputParameterError("get password in body")
		}
		// try base64 decryption
		if decPasswd, err := base64.StdEncoding.DecodeString(passwd); err == nil && stringutils2.IsPrintableAsciiString(string(decPasswd)) {
			passwd = string(decPasswd)
		}
		if len(uname) == 0 || len(passwd) == 0 {
			return nil, httperrors.NewInputParameterError("username or password is empty")
		}

		tenant, uname = parseLoginUser(uname)
		// var token mcclient.TokenCredential
		domain, _ := body.GetString("domain")
		token, err = auth.Client().AuthenticateWeb(uname, passwd, domain, "", "", cliIp)
	} else if body.Contains("idp_driver") { // sso login
		token, err = processSsoLoginData(body, cliIp)
		if err != nil {
			return nil, errors.Wrap(err, "processSsoLoginData")
		}
	} else {
		return nil, httperrors.NewInputParameterError("missing credential")
	}
	if err != nil {
		switch httperr := err.(type) {
		case *httputils.JSONClientError:
			if httperr.Code == 409 || httperr.Code == 429 {
				return nil, err
			}
		}
		return nil, httperrors.NewInvalidCredentialError("invalid credential")
	}
	uname := token.GetUserName()
	if len(tenant) > 0 {
		s := auth.GetAdminSession(ctx, FetchRegion(req), "")
		jsonProj, e := modules.Projects.GetById(s, tenant, nil)
		if e != nil {
			log.Errorf("fail to find preset project %s, reset to empty", tenant)
			tenant = ""
		} else {
			projId, _ := jsonProj.GetString("id")
			// projName, _ := jsonProj.GetString("name")
			ntoken, e := auth.Client().SetProject(projId, "", "", token)
			if e != nil {
				log.Errorf("fail to change to preset project %s(%s), reset to empty", tenant, e)
				tenant = ""
			} else {
				token = ntoken
			}
		}
	}
	if len(tenant) == 0 {
		token3, ok := token.(*mcclient.TokenCredentialV3)
		if ok {
			targetLevel := rbacutils.ScopeProject
			targetProjId := ""
			for _, r := range token3.Token.RoleAssignments {
				level := rbacutils.ScopeProject
				if len(r.Policies.System) > 0 {
					level = rbacutils.ScopeSystem
				} else if len(r.Policies.Domain) > 0 {
					level = rbacutils.ScopeDomain
				}
				if len(targetProjId) == 0 || level.HigherThan(targetLevel) {
					targetProjId = r.Scope.Project.Id
					targetLevel = level
				}
			}
			if len(targetProjId) > 0 {
				ntoken, e := auth.Client().SetProject(targetProjId, "", "", token)
				if e != nil {
					log.Errorf("fail to change to project %s(%s), reset to empty", targetProjId, e)
				} else {
					token = ntoken
					tenant = targetProjId
					body.(*jsonutils.JSONDict).Set("scope", jsonutils.NewString(string(targetLevel)))
				}
			}
		}
	}
	if len(tenant) == 0 {
		s := auth.GetAdminSession(ctx, FetchRegion(req), "")
		projects, e := modules.UsersV3.GetProjects(s, token.GetUserId())
		if e == nil && len(projects.Data) > 0 {
			projectJson := projects.Data[0]
			for _, pJson := range projects.Data {
				pname, _ := pJson.GetString("name")
				if pname == uname {
					projectJson = pJson
					break
				}
			}
			pid, e := projectJson.GetString("id")
			if e == nil {
				ntoken, e := auth.Client().SetProject(pid, "", "", token)
				if e == nil {
					token = ntoken
				} else {
					log.Errorf("fail to change to default project %s(%s), reset to empty", pid, e)
				}
			}
		} else {
			log.Errorf("GetProjects for login user error %s project count %d", e, len(projects.Data))
		}
	}
	return token, nil
}

func parseLoginUser(uname string) (string, string) {
	slashpos := strings.IndexByte(uname, '/')
	tenant := ""
	if slashpos > 0 {
		tenant = uname[0:slashpos]
		uname = uname[slashpos+1:]
	}

	return tenant, uname
}

func isUserAllowWebconsole(userInfo jsonutils.JSONObject) bool {
	return jsonutils.QueryBoolean(userInfo, "allow_web_console", true)
}

func saveCookie(w http.ResponseWriter, name, val, domain string, expire time.Time, base64 bool) {
	diff := time.Until(expire)
	maxAge := int(diff.Seconds())
	// log.Println("Set cookie", name, expire, maxAge, val)
	var valenc string
	if base64 {
		valenc = Base64UrlEncode([]byte(val))
	} else {
		valenc = val
	}
	// log.Printf("Set coookie: %s - %s\n", val, valenc)
	cookie := &http.Cookie{Name: name, Value: valenc, Path: "/", Expires: expire, MaxAge: maxAge, HttpOnly: false}

	if len(domain) > 0 {
		cookie.Domain = domain
	}

	http.SetCookie(w, cookie)
}

func getCookie(r *http.Request, name string) string {
	return getCookie2(r, name, true)
}

func getCookie2(r *http.Request, name string, base64 bool) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		log.Errorf("Cookie not found %q", name)
		return ""
		// } else if cookie.Expires.Before(time.Now()) {
		//     fmt.Println("Cookie expired ", cookie.Expires, time.Now())
		//     return ""
	} else {
		if !base64 {
			return cookie.Value
		}
		val, err := Base64UrlDecode(cookie.Value)
		if err != nil {
			log.Errorf("Cookie %q fail to decode: %v", name, err)
			return ""
		}
		return string(val)
	}
}

func clearCookie(w http.ResponseWriter, name string, domain string) {
	cookie := &http.Cookie{Name: name, Expires: time.Now(), Path: "/", MaxAge: -1, HttpOnly: false}
	if len(domain) > 0 {
		cookie.Domain = domain
	}

	http.SetCookie(w, cookie)
}

func saveAuthCookie(w http.ResponseWriter, authToken *clientman.SAuthToken, token mcclient.TokenCredential) {
	authCookie := authToken.GetAuthCookie(token)
	saveCookie(w, constants.YUNION_AUTH_COOKIE, authCookie, options.Options.CookieDomain, token.GetExpires(), true)
}

func clearAuthCookie(w http.ResponseWriter) {
	clearCookie(w, constants.YUNION_AUTH_COOKIE, options.Options.CookieDomain)
}

type PreLoginFunc func(ctx context.Context, req *http.Request, uname string, body jsonutils.JSONObject) error

func (h *AuthHandlers) postLoginHandler(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	body, err := appsrv.FetchJSON(req)
	if err != nil {
		httperrors.InvalidInputError(ctx, w, "fetch json for request: %v", err)
		return
	}
	err = h.doLogin(ctx, w, req, body)
	if err != nil {
		httperrors.GeneralServerError(ctx, w, err)
		return
	}
	// normal
	appsrv.Send(w, "")
}

func (h *AuthHandlers) doLogin(ctx context.Context, w http.ResponseWriter, req *http.Request, body jsonutils.JSONObject) error {
	var err error
	var authToken *clientman.SAuthToken
	var token mcclient.TokenCredential
	var userInfo jsonutils.JSONObject
	if body.Contains("tenantId") { // switch project
		token, authToken, err = doTenantLogin(ctx, req, body)
		if err != nil {
			return errors.Wrap(err, "doTenantLogin")
		}
		userInfo, err = fetchUserInfoFromToken(ctx, req, token)
		if err != nil {
			return errors.Wrap(err, "fetchUserInfoFromToken")
		}
	} else {
		// user/password authenticate
		// SSO authentication
		token, err = h.doCredentialLogin(ctx, req, body)
		if err != nil {
			return errors.Wrap(err, "doCredentialLogin")
		}
		userInfo, err = fetchUserInfoFromToken(ctx, req, token)
		if err != nil {
			return errors.Wrap(err, "fetchUserInfoFromToken")
		}
		s := auth.GetAdminSession(ctx, FetchRegion(req), "")
		isTotpInit, err := isUserTotpCredInitialed(s, token.GetUserId())
		if err != nil {
			return errors.Wrap(err, "isUserTotpCredInitialed")
		}
		authToken = clientman.NewAuthToken(token.GetTokenString(), isUserEnableTotp(userInfo), isTotpInit)
	}

	if !isUserAllowWebconsole(userInfo) {
		return errors.Wrap(httperrors.ErrForbidden, "user forbidden login from web")
	}

	saveAuthCookie(w, authToken, token)

	if len(token.GetProjectId()) > 0 {
		if body.Contains("isadmin") {
			adminVal := "false"
			if policy.PolicyManager.IsScopeCapable(token, rbacutils.ScopeSystem) {
				adminVal, _ = body.GetString("isadmin")
			}
			saveCookie(w, "isadmin", adminVal, "", token.GetExpires(), false)
		}
		if body.Contains("scope") {
			scopeStr, _ := body.GetString("scope")
			if !policy.PolicyManager.IsScopeCapable(token, rbacutils.TRbacScope(scopeStr)) {
				scopeStr = string(rbacutils.ScopeProject)
			}
			saveCookie(w, "scope", scopeStr, "", token.GetExpires(), false)
		}
		if body.Contains("domain") {
			domainStr, _ := body.GetString("domain")
			saveCookie(w, "domain", domainStr, "", token.GetExpires(), false)
		}
		saveCookie(w, "tenant", token.GetProjectId(), "", token.GetExpires(), false)
	}

	return nil
}

func (h *AuthHandlers) postLogoutHandler(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	clearAuthCookie(w)
	appsrv.Send(w, "")
}

func FetchRegion(req *http.Request) string {
	r, e := req.Cookie("region")
	if e == nil && len(r.Value) > 0 {
		return r.Value
	}
	if len(options.Options.DefaultRegion) > 0 {
		return options.Options.DefaultRegion
	}
	adminToken := auth.AdminCredential()
	if adminToken == nil {
		log.Errorf("FetchRegion: nil adminTken")
		return ""
	}
	regions := adminToken.GetRegions()
	if len(regions) == 0 {
		log.Errorf("FetchRegion: empty region list")
		return ""
	}
	for _, r := range regions {
		if len(r) > 0 {
			return r
		}
	}
	log.Errorf("FetchRegion: no valid region")
	return ""
}

type role struct {
	id   string
	name string
}

type projectRoles struct {
	id       string
	name     string
	domain   string
	domainId string
	roles    []role
}

func newProjectRoles(projectId, projectName, roleId, roleName string, domainId, domainName string) *projectRoles {
	return &projectRoles{
		id:       projectId,
		name:     projectName,
		domainId: domainId,
		domain:   domainName,
		roles:    []role{{id: roleId, name: roleName}},
	}
}

func (this *projectRoles) add(roleId, roleName string) {
	this.roles = append(this.roles, role{id: roleId, name: roleName})
}

func (this *projectRoles) getToken(scope rbacutils.TRbacScope, user, userId, domain, domainId string, ip string) mcclient.TokenCredential {
	return &mcclient.SSimpleToken{
		Token:           "faketoken",
		Domain:          domain,
		DomainId:        domainId,
		User:            user,
		UserId:          userId,
		Project:         this.name,
		ProjectId:       this.id,
		ProjectDomain:   this.domain,
		ProjectDomainId: this.domainId,
		Roles:           strings.Join(this.getRoles(), ","),
		Context: mcclient.SAuthContext{
			Ip: ip,
		},
	}
	// return policy.PolicyManager.IsScopeCapable(&t, scope)
}

func (this *projectRoles) getRoles() []string {
	roles := make([]string, 0)
	for _, r := range this.roles {
		roles = append(roles, r.name)
	}
	return roles
}

func (this *projectRoles) json(user, userId, domain, domainId string, ip string) jsonutils.JSONObject {
	obj := jsonutils.NewDict()
	obj.Add(jsonutils.NewString(this.id), "id")
	obj.Add(jsonutils.NewString(this.name), "name")
	obj.Add(jsonutils.NewString(this.domain), "domain")
	obj.Add(jsonutils.NewString(this.domainId), "domain_id")
	roles := jsonutils.NewArray()
	for _, r := range this.roles {
		role := jsonutils.NewDict()
		role.Add(jsonutils.NewString(r.id), "id")
		role.Add(jsonutils.NewString(r.name), "name")
		roles.Add(role)
	}
	obj.Add(roles, "roles")
	for _, scope := range []rbacutils.TRbacScope{
		rbacutils.ScopeProject,
		rbacutils.ScopeDomain,
		rbacutils.ScopeSystem,
	} {
		token := this.getToken(scope, user, userId, domain, domainId, ip)
		matches := policy.PolicyManager.MatchedPolicyNames(scope, token)
		obj.Add(jsonutils.NewStringArray(matches), fmt.Sprintf("%s_policies", scope))
		if len(matches) > 0 {
			obj.Add(jsonutils.JSONTrue, fmt.Sprintf("%s_capable", scope))
		} else {
			obj.Add(jsonutils.JSONFalse, fmt.Sprintf("%s_capable", scope))
		}
		// backward compatible
		if scope == rbacutils.ScopeSystem {
			if len(matches) > 0 {
				obj.Add(jsonutils.JSONTrue, "admin_capable")
			} else {
				obj.Add(jsonutils.JSONFalse, "admin_capable")
			}
		}
	}
	return obj
}

func isLBAgentExists(s *mcclient.ClientSession) (bool, error) {
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString("hb_last_seen.isnotempty()"), "filter.0")
	params.Add(jsonutils.NewInt(1), "limit")
	params.Add(jsonutils.JSONFalse, "details")
	agents, err := modules.LoadbalancerAgents.List(s, params)
	if err != nil {
		return false, errors.Wrap(err, "modules.LoadbalancerAgents.List")
	}

	if len(agents.Data) > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func isBaremetalAgentExists(s *mcclient.ClientSession) (bool, error) {
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString("agent_type.equals(baremetal)"), "filter.0")
	params.Add(jsonutils.NewInt(1), "limit")
	params.Add(jsonutils.JSONFalse, "details")
	agents, err := modules.Baremetalagents.List(s, params)
	if err != nil {
		return false, errors.Wrap(err, "modules.Baremetalagents.List")
	}

	if len(agents.Data) > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func isEsxiAgentExists(s *mcclient.ClientSession) (bool, error) {
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString("agent_type.equals(esxiagent)"), "filter.0")
	params.Add(jsonutils.NewInt(1), "limit")
	params.Add(jsonutils.JSONFalse, "details")
	agents, err := modules.Baremetalagents.List(s, params)
	if err != nil {
		return false, errors.Wrap(err, "modules.Baremetalagents.List esxiagent")
	}

	if len(agents.Data) > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func isHostAgentExists(s *mcclient.ClientSession) (bool, error) {
	params := jsonutils.NewDict()
	params.Add(jsonutils.JSONFalse, "show_emulated")
	params.Add(jsonutils.JSONFalse, "baremetal")
	params.Add(jsonutils.NewString("system"), "scope")
	params.Add(jsonutils.NewInt(1), "limit")
	params.Add(jsonutils.JSONFalse, "details")
	agents, err := modules.Hosts.List(s, params)
	if err != nil {
		return false, errors.Wrap(err, "modules.LoadbalancerAgents.List")
	}

	if len(agents.Data) > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func getUserInfo(ctx context.Context, req *http.Request) (*jsonutils.JSONDict, error) {
	token := AppContextToken(ctx)
	s := auth.GetAdminSession(ctx, FetchRegion(req), "")
	/*log.Infof("getUserInfo modules.UsersV3.Get")
	usr, err := modules.UsersV3.Get(s, token.GetUserId(), nil)
	if err != nil {
		log.Errorf("modules.UsersV3.Get fail %s", err)
		return nil, fmt.Errorf("not found user %s", token.GetUserId())
	}*/
	usr, err := fetchUserInfoFromToken(ctx, req, token)
	if err != nil {
		return nil, errors.Wrapf(err, "fetchUserInfoFromToken %s", token.GetUserId())
	}
	data := jsonutils.NewDict()
	for _, k := range []string{
		"displayname", "email", "id", "name",
		"enabled", "mobile", "allow_web_console",
		"created_at", "enable_mfa", "is_system_account",
		"last_active_at", "last_login_ip",
		"last_login_source",
		"password_expires_at", "failed_auth_count", "failed_auth_at",
		"idps",
	} {
		v, e := usr.Get(k)
		if e == nil {
			data.Add(v, k)
		}
	}
	data.Add(jsonutils.NewString(token.GetDomainId()), "domain", "id")
	data.Add(jsonutils.NewString(token.GetDomainName()), "domain", "name")
	data.Add(jsonutils.NewStringArray(auth.AdminCredential().GetRegions()), "regions")
	data.Add(jsonutils.NewStringArray(token.GetRoles()), "roles")
	data.Add(jsonutils.NewString(token.GetProjectName()), "projectName")
	data.Add(jsonutils.NewString(token.GetProjectId()), "projectId")
	data.Add(jsonutils.NewString(token.GetProjectDomain()), "projectDomain")
	data.Add(jsonutils.NewString(token.GetProjectDomainId()), "projectDomainId")

	log.Infof("getUserInfo modules.RoleAssignments.List")
	query := jsonutils.NewDict()
	query.Add(jsonutils.JSONNull, "effective")
	query.Add(jsonutils.JSONNull, "include_names")
	query.Add(jsonutils.JSONNull, "include_system")
	query.Add(jsonutils.NewInt(0), "limit")
	query.Add(jsonutils.NewString(token.GetUserId()), "user", "id")
	roleAssigns, err := modules.RoleAssignments.List(s, query)
	if err != nil {
		return nil, errors.Wrapf(err, "get RoleAssignments list")
	}
	projects := make(map[string]*projectRoles)
	for _, roleAssign := range roleAssigns.Data {
		roleId, _ := roleAssign.GetString("role", "id")
		roleName, _ := roleAssign.GetString("role", "name")
		projectId, _ := roleAssign.GetString("scope", "project", "id")
		projectName, _ := roleAssign.GetString("scope", "project", "name")
		domainId, _ := roleAssign.GetString("scope", "project", "domain", "id")
		domain, _ := roleAssign.GetString("scope", "project", "domain", "name")
		_, ok := projects[projectId]
		if ok {
			projects[projectId].add(roleId, roleName)
		} else {
			projects[projectId] = newProjectRoles(projectId, projectName, roleId, roleName, domainId, domain)
		}
	}
	projJson := jsonutils.NewArray()
	for _, proj := range projects {
		projJson.Add(proj.json(
			token.GetUserName(),
			token.GetUserId(),
			token.GetDomainName(),
			token.GetDomainId(),
			token.GetLoginIp(),
		))
	}
	data.Add(projJson, "projects")

	for _, scope := range []rbacutils.TRbacScope{
		rbacutils.ScopeSystem,
		rbacutils.ScopeDomain,
		rbacutils.ScopeProject,
	} {
		p := policy.PolicyManager.MatchedPolicyNames(scope, token)
		data.Add(jsonutils.NewStringArray(p), fmt.Sprintf("%s_policies", scope))
		if scope == rbacutils.ScopeSystem {
			data.Add(jsonutils.NewStringArray(p), "admin_policies")
		} else if scope == rbacutils.ScopeProject {
			data.Add(jsonutils.NewStringArray(p), "policies")
		}
	}
	allPolicies := policy.PolicyManager.AllPolicies()
	data.Add(jsonutils.Marshal(allPolicies), "all_policies")

	services := jsonutils.NewArray()
	menus := jsonutils.NewArray()
	k8s := jsonutils.NewArray()

	curReg := FetchRegion(req)
	srvCat := auth.Client().GetServiceCatalog()
	var allsrv []string
	var alleps []mcclient.ExternalService
	if srvCat != nil {
		allsrv = srvCat.GetInternalServices(curReg)
		alleps = srvCat.GetServicesByInterface(curReg, "console")
	}

	log.Infof("getUserInfo checkAgent exists")
	for _, cf := range []struct {
		existFunc func(*mcclient.ClientSession) (bool, error)
		srvName   string
	}{
		{
			existFunc: isLBAgentExists,
			srvName:   "lbagent",
		},
		{
			existFunc: isBaremetalAgentExists,
			srvName:   "bmagent",
		},
		{
			existFunc: isHostAgentExists,
			srvName:   "hostagent",
		},
		{
			existFunc: isEsxiAgentExists,
			srvName:   "esxiagent",
		},
	} {
		exist, err := cf.existFunc(s)
		if err != nil {
			log.Errorf("isLBAgentExists fail %s", err)
		} else if exist {
			allsrv = append(allsrv, cf.srvName)
		}
	}

	for _, srv := range allsrv {
		item := jsonutils.NewDict()
		item.Add(jsonutils.NewString(srv), "type")
		item.Add(jsonutils.JSONTrue, "status")
		services.Add(item)
	}

	for _, ep := range alleps {
		item := jsonutils.NewDict()
		item.Add(jsonutils.NewString(ep.Url), "url")
		item.Add(jsonutils.NewString(ep.Name), "name")
		menus.Add(item)
	}

	log.Infof("getUserInfo modules.Hosts.Get")
	s2 := auth.GetSession(ctx, token, FetchRegion(req), "v2")
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString("host_type"), "field")
	params.Add(jsonutils.NewString("system"), "scope")
	params.Add(jsonutils.JSONTrue, "usable")
	params.Add(jsonutils.JSONTrue, "show_emulated")
	cap, err := modules.Hosts.Get(s2, "distinct-field", params)
	if err != nil {
		log.Errorf("modules.Servers.Get distinct-field fail %s", err)
	} else {
		hostTypes, _ := jsonutils.GetStringArray(cap, "host_type")
		hypervisors := make([]string, len(hostTypes))
		for i, hostType := range hostTypes {
			hypervisors[i] = compute.HOSTTYPE_HYPERVISOR[hostType]
		}
		data.Add(jsonutils.NewStringArray(hypervisors), "hypervisors")
	}

	data.Add(menus, "menus")
	data.Add(k8s, "k8sdashboard")
	data.Add(services, "services")

	if options.Options.NonDefaultDomainProjects {
		data.Add(jsonutils.JSONTrue, "non_default_domain_projects")
	} else {
		data.Add(jsonutils.JSONFalse, "non_default_domain_projects")
	}

	data.Add(jsonutils.NewString(getSsoCallbackUrl()), "sso_callback_url")

	return data, nil
}

func (h *AuthHandlers) getPermissionDetails(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	t := AppContextToken(ctx)

	_, query, body := appsrv.FetchEnv(ctx, w, req)
	if body == nil {
		httperrors.InvalidInputError(ctx, w, "body is empty")
		return
	}
	var name string
	if query != nil {
		name, _ = query.GetString("policy")
	}
	result, err := policy.ExplainRpc(ctx, t, body, name)
	if err != nil {
		httperrors.GeneralServerError(ctx, w, err)
		return
	}

	appsrv.SendJSON(w, result)
}

func (h *AuthHandlers) getAdminResources(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	res := policy.GetSystemResources()
	appsrv.SendJSON(w, jsonutils.Marshal(res))
}

func (h *AuthHandlers) getResources(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	res := policy.GetResources()
	appsrv.SendJSON(w, jsonutils.Marshal(res))
}

func (h *AuthHandlers) doCreatePolicies(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	t := AppContextToken(ctx)
	// if !utils.IsInStringArray("admin", t.GetRoles()) || t.GetProjectName() != "system" {
	// 	httperrors.ForbiddenError(ctx, w, "not allow to create policy")
	// 	return
	// }
	_, _, body := appsrv.FetchEnv(ctx, w, req)
	if body == nil {
		httperrors.InvalidInputError(ctx, w, "body is empty")
		return
	}
	s := auth.GetSession(ctx, t, FetchRegion(req), "")
	result, err := policytool.PolicyCreate(s, body)
	if err != nil {
		httperrors.GeneralServerError(ctx, w, err)
		return
	}
	appsrv.SendJSON(w, result)
}

func (h *AuthHandlers) doPatchPolicy(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	t := AppContextToken(ctx)
	// if !utils.IsInStringArray("admin", t.GetRoles()) || t.GetProjectName() != "system" {
	// 	httperrors.ForbiddenError(ctx, w, "not allow to create policy")
	// 	return
	// }
	params, _, body := appsrv.FetchEnv(ctx, w, req)
	if body == nil {
		httperrors.InvalidInputError(ctx, w, "request body is empty")
		return
	}
	s := auth.GetSession(ctx, t, FetchRegion(req), "")
	result, err := policytool.PolicyPatch(s, params["<policy_id>"], body)
	if err != nil {
		httperrors.GeneralServerError(ctx, w, err)
		return
	}
	appsrv.SendJSON(w, result)
}

func (h *AuthHandlers) doDeletePolicies(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	t := AppContextToken(ctx)
	// if !utils.IsInStringArray("admin", t.GetRoles()) || t.GetProjectName() != "system" {
	// 	httperrors.ForbiddenError(ctx, w, "not allow to create policy")
	// 	return
	// }
	_, query, _ := appsrv.FetchEnv(ctx, w, req)
	s := auth.GetSession(ctx, t, FetchRegion(req), "")

	idlist, e := query.GetArray("id")
	if e != nil || len(idlist) == 0 {
		httperrors.InvalidInputError(ctx, w, "missing id")
		return
	}
	idStrList := jsonutils.JSONArray2StringArray(idlist)
	ret := make([]modulebase.SubmitResult, len(idStrList))
	for i := range idStrList {
		err := policytool.PolicyDelete(s, idStrList[i])
		if err != nil {
			ret[i] = modulebase.SubmitResult{
				Status: 400,
				Id:     idStrList[i],
				Data:   jsonutils.NewString(err.Error()),
			}
		} else {
			ret[i] = modulebase.SubmitResult{
				Status: 200,
				Id:     idStrList[i],
				Data:   jsonutils.NewDict(),
			}
		}
	}
	w.WriteHeader(207)
	appsrv.SendJSON(w, modulebase.SubmitResults2JSON(ret))
}

/*
重置密码
1.验证新密码正确
2.验证原密码正确，且idp_driver为空
3.如果已开启MFA，验证 随机密码正确
4.重置密码，清除认证token
*/
func (h *AuthHandlers) resetUserPassword(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	t, authToken, err := fetchAuthInfo(ctx, req)
	if err != nil {
		httperrors.InvalidCredentialError(ctx, w, "fetchAuthInfo fail: %s", err)
		return
	}

	_, _, body := appsrv.FetchEnv(ctx, w, req)
	if body == nil {
		httperrors.InvalidInputError(ctx, w, "body is empty")
		return
	}

	user, err := fetchUserInfoFromToken(ctx, req, t)
	if err != nil {
		httperrors.GeneralServerError(ctx, w, err)
		return
	}

	oldPwd, _ := body.GetString("password_old")
	newPwd, _ := body.GetString("password_new")
	confirmPwd, _ := body.GetString("password_confirm")
	passcode, _ := body.GetString("passcode")

	if newPwd != confirmPwd {
		httperrors.InputParameterError(ctx, w, "new password mismatch")
		return
	}

	// 1.验证原密码正确，且idp_driver为空
	if isIdpUser(user) {
		httperrors.ForbiddenError(ctx, w, "not support reset user password")
		return
	}

	cliIp := netutils2.GetHttpRequestIp(req)
	_, err = auth.Client().AuthenticateWeb(t.GetUserName(), oldPwd, t.GetDomainName(), "", "", cliIp)
	if err != nil {
		switch httperr := err.(type) {
		case *httputils.JSONClientError:
			if httperr.Code == 409 {
				httperrors.GeneralServerError(ctx, w, err)
				return
			}
		}
		httperrors.InputParameterError(ctx, w, "密码错误")
		return
	}

	s := auth.GetAdminSession(ctx, FetchRegion(req), "")
	// 2.如果已开启MFA，验证 随机密码正确
	if isMfaEnabled(user) {
		err = authToken.VerifyTotpPasscode(s, t.GetUserId(), passcode)
		if err != nil {
			httperrors.InputParameterError(ctx, w, "invalid passcode")
			return
		}
	}

	// 3.重置密码，
	params := jsonutils.NewDict()
	params.Set("password", jsonutils.NewString(newPwd))
	_, err = modules.UsersV3.Patch(s, t.GetUserId(), params)
	if err != nil {
		httperrors.GeneralServerError(ctx, w, err)
		return
	}

	// 4. 清除认证token, logout
	h.postLogoutHandler(ctx, w, req)
}

func isIdpUser(user jsonutils.JSONObject) bool {
	if driver, _ := user.GetString("idp_driver"); len(driver) > 0 {
		return true
	}

	return false
}

// refer: isUserEnableTotp
func isMfaEnabled(user jsonutils.JSONObject) bool {
	if !options.Options.EnableTotp {
		return false
	}

	if ok, _ := user.Bool("enable_mfa"); ok {
		return true
	}

	return false
}
