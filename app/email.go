// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"fmt"
	"io"
	"net/url"
	"path"

	"net/http"

	"github.com/mattermost/go-i18n/i18n"
	"github.com/pkg/errors"
	"github.com/throttled/throttled"
	"github.com/throttled/throttled/store/memstore"

	"github.com/nhannv/quiz/v5/model"
	"github.com/nhannv/quiz/v5/services/mailservice"
	"github.com/nhannv/quiz/v5/utils"
)

const (
	emailRateLimitingMemstoreSize = 65536
	emailRateLimitingPerHour      = 20
	emailRateLimitingMaxBurst     = 20
)

func condenseSiteURL(siteURL string) string {
	parsedSiteURL, _ := url.Parse(siteURL)
	if parsedSiteURL.Path == "" || parsedSiteURL.Path == "/" {
		return parsedSiteURL.Host
	}

	return path.Join(parsedSiteURL.Host, parsedSiteURL.Path)
}

func (a *App) SetupInviteEmailRateLimiting() error {
	store, err := memstore.New(emailRateLimitingMemstoreSize)
	if err != nil {
		return errors.Wrap(err, "Unable to setup email rate limiting memstore.")
	}

	quota := throttled.RateQuota{
		MaxRate:  throttled.PerHour(emailRateLimitingPerHour),
		MaxBurst: emailRateLimitingMaxBurst,
	}

	rateLimiter, err := throttled.NewGCRARateLimiter(store, quota)
	if err != nil || rateLimiter == nil {
		return errors.Wrap(err, "Unable to setup email rate limiting GCRA rate limiter.")
	}

	a.Srv().EmailRateLimiter = rateLimiter
	return nil
}

func (a *App) SendChangeUsernameEmail(oldUsername, newUsername, email, locale, siteURL string) *model.AppError {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.username_change_subject",
		map[string]interface{}{"SiteName": a.ClientConfig()["SiteName"],
			"TeamDisplayName": a.Config().TeamSettings.SiteName})

	bodyPage := a.NewEmailTemplate("email_change_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.username_change_body.title")
	bodyPage.Props["Info"] = T("api.templates.username_change_body.info",
		map[string]interface{}{"TeamDisplayName": a.Config().TeamSettings.SiteName, "NewUsername": newUsername})
	bodyPage.Props["Warning"] = T("api.templates.email_warning")

	if err := a.SendMail(email, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("SendChangeUsernameEmail", "api.user.send_email_change_username_and_forget.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) SendEmailChangeVerifyEmail(newUserEmail, locale, siteURL, token string) *model.AppError {
	T := utils.GetUserTranslations(locale)

	link := fmt.Sprintf("%s/do_verify_email?token=%s&email=%s", siteURL, token, url.QueryEscape(newUserEmail))

	subject := T("api.templates.email_change_verify_subject",
		map[string]interface{}{"SiteName": a.ClientConfig()["SiteName"],
			"TeamDisplayName": a.Config().TeamSettings.SiteName})

	bodyPage := a.NewEmailTemplate("email_change_verify_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.email_change_verify_body.title")
	bodyPage.Props["Info"] = T("api.templates.email_change_verify_body.info",
		map[string]interface{}{"TeamDisplayName": a.Config().TeamSettings.SiteName})
	bodyPage.Props["VerifyUrl"] = link
	bodyPage.Props["VerifyButton"] = T("api.templates.email_change_verify_body.button")

	if err := a.SendMail(newUserEmail, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("SendEmailChangeVerifyEmail", "api.user.send_email_change_verify_email_and_forget.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) SendEmailChangeEmail(oldEmail, newEmail, locale, siteURL string) *model.AppError {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.email_change_subject",
		map[string]interface{}{"SiteName": a.ClientConfig()["SiteName"],
			"TeamDisplayName": a.Config().TeamSettings.SiteName})

	bodyPage := a.NewEmailTemplate("email_change_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.email_change_body.title")
	bodyPage.Props["Info"] = T("api.templates.email_change_body.info",
		map[string]interface{}{"TeamDisplayName": a.Config().TeamSettings.SiteName, "NewEmail": newEmail})
	bodyPage.Props["Warning"] = T("api.templates.email_warning")

	if err := a.SendMail(oldEmail, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("SendEmailChangeEmail", "api.user.send_email_change_email_and_forget.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) SendVerifyEmail(userEmail, locale, siteURL, token string) *model.AppError {
	T := utils.GetUserTranslations(locale)

	link := fmt.Sprintf("%s/do_verify_email?token=%s&email=%s", siteURL, token, url.QueryEscape(userEmail))

	serverURL := condenseSiteURL(siteURL)

	subject := T("api.templates.verify_subject",
		map[string]interface{}{"SiteName": a.ClientConfig()["SiteName"]})

	bodyPage := a.NewEmailTemplate("verify_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.verify_body.title", map[string]interface{}{"ServerURL": serverURL})
	bodyPage.Props["Info"] = T("api.templates.verify_body.info")
	bodyPage.Props["VerifyUrl"] = link
	bodyPage.Props["Button"] = T("api.templates.verify_body.button")

	if err := a.SendMail(userEmail, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("SendVerifyEmail", "api.user.send_verify_email_and_forget.failed.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) SendSignInChangeEmail(email, method, locale, siteURL string) *model.AppError {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.signin_change_email.subject",
		map[string]interface{}{"SiteName": a.ClientConfig()["SiteName"]})

	bodyPage := a.NewEmailTemplate("signin_change_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.signin_change_email.body.title")
	bodyPage.Props["Info"] = T("api.templates.signin_change_email.body.info",
		map[string]interface{}{"SiteName": a.ClientConfig()["SiteName"], "Method": method})
	bodyPage.Props["Warning"] = T("api.templates.email_warning")

	if err := a.SendMail(email, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("SendSignInChangeEmail", "api.user.send_sign_in_change_email_and_forget.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) SendWelcomeEmail(userId string, email string, verified bool, locale, siteURL string) *model.AppError {
	if !*a.Config().EmailSettings.SendEmailNotifications && !*a.Config().EmailSettings.RequireEmailVerification {
		return model.NewAppError("SendWelcomeEmail", "api.user.send_welcome_email_and_forget.failed.error", nil, "Send Email Notifications and Require Email Verification is disabled in the system console", http.StatusInternalServerError)
	}

	T := utils.GetUserTranslations(locale)

	serverURL := condenseSiteURL(siteURL)

	subject := T("api.templates.welcome_subject",
		map[string]interface{}{"SiteName": a.ClientConfig()["SiteName"],
			"ServerURL": serverURL})

	bodyPage := a.NewEmailTemplate("welcome_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.welcome_body.title", map[string]interface{}{"ServerURL": serverURL})
	bodyPage.Props["Info"] = T("api.templates.welcome_body.info")
	bodyPage.Props["Button"] = T("api.templates.welcome_body.button")
	bodyPage.Props["Info2"] = T("api.templates.welcome_body.info2")
	bodyPage.Props["Info3"] = T("api.templates.welcome_body.info3")
	bodyPage.Props["SiteURL"] = siteURL

	if *a.Config().NativeAppSettings.AppDownloadLink != "" {
		bodyPage.Props["AppDownloadInfo"] = T("api.templates.welcome_body.app_download_info")
		bodyPage.Props["AppDownloadLink"] = *a.Config().NativeAppSettings.AppDownloadLink
	}

	if !verified && *a.Config().EmailSettings.RequireEmailVerification {
		token, err := a.CreateVerifyEmailToken(userId, email)
		if err != nil {
			return err
		}
		link := fmt.Sprintf("%s/do_verify_email?token=%s&email=%s", siteURL, token.Token, url.QueryEscape(email))
		bodyPage.Props["VerifyUrl"] = link
	}

	if err := a.SendMail(email, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("SendWelcomeEmail", "api.user.send_welcome_email_and_forget.failed.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) SendPasswordChangeEmail(email, method, locale, siteURL string) *model.AppError {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.password_change_subject",
		map[string]interface{}{"SiteName": a.ClientConfig()["SiteName"],
			"TeamDisplayName": a.Config().TeamSettings.SiteName})

	bodyPage := a.NewEmailTemplate("password_change_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.password_change_body.title")
	bodyPage.Props["Info"] = T("api.templates.password_change_body.info",
		map[string]interface{}{"TeamDisplayName": a.Config().TeamSettings.SiteName, "TeamURL": siteURL, "Method": method})
	bodyPage.Props["Warning"] = T("api.templates.email_warning")

	if err := a.SendMail(email, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("SendPasswordChangeEmail", "api.user.send_password_change_email_and_forget.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) SendUserAccessTokenAddedEmail(email, locale, siteURL string) *model.AppError {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.user_access_token_subject",
		map[string]interface{}{"SiteName": a.ClientConfig()["SiteName"]})

	bodyPage := a.NewEmailTemplate("password_change_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.user_access_token_body.title")
	bodyPage.Props["Info"] = T("api.templates.user_access_token_body.info",
		map[string]interface{}{"SiteName": a.ClientConfig()["SiteName"], "SiteURL": siteURL})
	bodyPage.Props["Warning"] = T("api.templates.email_warning")

	if err := a.SendMail(email, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("SendUserAccessTokenAddedEmail", "api.user.send_user_access_token.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) SendPasswordResetEmail(email string, token *model.Token, locale, siteURL string) (bool, *model.AppError) {
	T := utils.GetUserTranslations(locale)

	link := fmt.Sprintf("%s/reset_password_complete?token=%s", siteURL, url.QueryEscape(token.Token))

	subject := T("api.templates.reset_subject",
		map[string]interface{}{"SiteName": a.ClientConfig()["SiteName"]})

	bodyPage := a.NewEmailTemplate("reset_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.reset_body.title")
	bodyPage.Props["Info1"] = utils.TranslateAsHtml(T, "api.templates.reset_body.info1", nil)
	bodyPage.Props["Info2"] = T("api.templates.reset_body.info2")
	bodyPage.Props["ResetUrl"] = link
	bodyPage.Props["Button"] = T("api.templates.reset_body.button")

	if err := a.SendMail(email, subject, bodyPage.Render()); err != nil {
		return false, model.NewAppError("SendPasswordReset", "api.user.send_password_reset.send.app_error", nil, "err="+err.Message, http.StatusInternalServerError)
	}

	return true, nil
}

func (a *App) SendMfaChangeEmail(email string, activated bool, locale, siteURL string) *model.AppError {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.mfa_change_subject",
		map[string]interface{}{"SiteName": a.ClientConfig()["SiteName"]})

	bodyPage := a.NewEmailTemplate("mfa_change_body", locale)
	bodyPage.Props["SiteURL"] = siteURL

	if activated {
		bodyPage.Props["Info"] = T("api.templates.mfa_activated_body.info", map[string]interface{}{"SiteURL": siteURL})
		bodyPage.Props["Title"] = T("api.templates.mfa_activated_body.title")
	} else {
		bodyPage.Props["Info"] = T("api.templates.mfa_deactivated_body.info", map[string]interface{}{"SiteURL": siteURL})
		bodyPage.Props["Title"] = T("api.templates.mfa_deactivated_body.title")
	}
	bodyPage.Props["Warning"] = T("api.templates.email_warning")

	if err := a.SendMail(email, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("SendMfaChangeEmail", "api.user.send_mfa_change_email.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) NewEmailTemplate(name, locale string) *utils.HTMLTemplate {
	t := utils.NewHTMLTemplate(a.HTMLTemplates(), name)

	var localT i18n.TranslateFunc
	if locale != "" {
		localT = utils.GetUserTranslations(locale)
	} else {
		localT = utils.T
	}

	t.Props["Footer"] = localT("api.templates.email_footer")

	if *a.Config().EmailSettings.FeedbackOrganization != "" {
		t.Props["Organization"] = localT("api.templates.email_organization") + *a.Config().EmailSettings.FeedbackOrganization
	} else {
		t.Props["Organization"] = ""
	}

	t.Props["EmailInfo1"] = localT("api.templates.email_info1")
	t.Props["EmailInfo2"] = localT("api.templates.email_info2")
	t.Props["EmailInfo3"] = localT("api.templates.email_info3",
		map[string]interface{}{"SiteName": a.Config().TeamSettings.SiteName})
	t.Props["SupportEmail"] = *a.Config().SupportSettings.SupportEmail

	return t
}

func (a *App) SendDeactivateAccountEmail(email string, locale, siteURL string) *model.AppError {
	T := utils.GetUserTranslations(locale)

	serverURL := condenseSiteURL(siteURL)

	subject := T("api.templates.deactivate_subject",
		map[string]interface{}{"SiteName": a.ClientConfig()["SiteName"],
			"ServerURL": serverURL})

	bodyPage := a.NewEmailTemplate("deactivate_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.deactivate_body.title", map[string]interface{}{"ServerURL": serverURL})
	bodyPage.Props["Info"] = T("api.templates.deactivate_body.info",
		map[string]interface{}{"SiteURL": siteURL})
	bodyPage.Props["Warning"] = T("api.templates.deactivate_body.warning")

	if err := a.SendMail(email, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("SendDeactivateEmail", "api.user.send_deactivate_email_and_forget.failed.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) SendNotificationMail(to, subject, htmlBody string) *model.AppError {
	if !*a.Config().EmailSettings.SendEmailNotifications {
		return nil
	}
	return a.SendMail(to, subject, htmlBody)
}

func (a *App) SendMail(to, subject, htmlBody string) *model.AppError {
	license := a.License()
	return mailservice.SendMailUsingConfig(to, subject, htmlBody, a.Config(), license != nil && *license.Features.Compliance)
}

func (a *App) SendMailWithEmbeddedFiles(to, subject, htmlBody string, embeddedFiles map[string]io.Reader) *model.AppError {
	license := a.License()
	config := a.Config()

	return mailservice.SendMailWithEmbeddedFilesUsingConfig(to, subject, htmlBody, embeddedFiles, config, license != nil && *license.Features.Compliance)
}
