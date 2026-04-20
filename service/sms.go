package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// Constants and configs
const (
	MaxSmsVerifyCodeFailedCount = 3
	SmsVerifyCodeFailedLockTime = 15 * time.Minute
	DefaultConnectTimeout       = 10 * time.Second
	VerifyCodeExpiration        = 10 * time.Minute
)

// TODO: Replace with your actual configuration retrieval mechanism
// Configs will be loaded from common

var (
	ErrVerifyCodeMaxTries = errors.New("verify code max tries exceeded or locked")
	ErrVerifyCodeNotUsed  = errors.New("verify code has been sent but not used")
	ErrVerifyCodeInvalid  = errors.New("invalid verify code")
	ErrSmsSendFailed      = errors.New("failed to send sms")
)

// SmsSender defines the SMS verification and sending interface
type SmsSender interface {
	SendCode(ctx context.Context, phoneNumber string) error
	SendSms(ctx context.Context, phoneNumber string, message string) error
	CheckSms(ctx context.Context, phoneNumber string, code string) (bool, error)
	DeleteSms(ctx context.Context, phoneNumber string) error
}

// BaseSmsSender implements common verification logic
type BaseSmsSender struct {
	Impl SmsSender // Used to call the specific provider's SendSms implementation
}

func (s *BaseSmsSender) SendCode(ctx context.Context, phoneNumber string) error {
	failedCountKey := fmt.Sprintf("sms_verify_code_failed_count:%s", phoneNumber)

	failedCountStr, err := common.RedisGet(failedCountKey)
	if err == nil && failedCountStr != "" {
		count, _ := strconv.Atoi(failedCountStr)
		if count >= MaxSmsVerifyCodeFailedCount {
			return fmt.Errorf("%w: locked for %v", ErrVerifyCodeMaxTries, SmsVerifyCodeFailedLockTime)
		}
	}

	sendedCodeKey := fmt.Sprintf("sms_verify_code:%s", phoneNumber)
	sendedCode, err := common.RedisGet(sendedCodeKey)
	if err == nil && sendedCode != "" {
		return ErrVerifyCodeNotUsed
	}

	code := s.RandCode(6)
	common.SysLog(fmt.Sprintf("verify code: %s", code))
	message := fmt.Sprintf("验证码：%s, 你正在使用短信验证码登录，有效期10分钟，请勿泄露", code)

	err = common.RedisSet(sendedCodeKey, code, VerifyCodeExpiration)
	if err != nil {
		return err
	}
	common.RedisDel(failedCountKey)

	// Since we compose, we must call the Impl's SendSms
	return s.Impl.SendSms(ctx, phoneNumber, message)
}

func (s *BaseSmsSender) CheckSms(ctx context.Context, phoneNumber string, code string) (bool, error) {
	if common.SMSWhitePhoneList != "" {
		whiteList := strings.Split(common.SMSWhitePhoneList, ",")
		for _, whitePhone := range whiteList {
			if phoneNumber == whitePhone && code == "aabbcc" {
				return true, nil
			}
		}
	}

	failedCountKey := fmt.Sprintf("sms_verify_code_failed_count:%s", phoneNumber)
	failedCountStr, err := common.RedisGet(failedCountKey)
	if err == nil && failedCountStr != "" {
		count, _ := strconv.Atoi(failedCountStr)
		if count >= MaxSmsVerifyCodeFailedCount {
			return false, fmt.Errorf("%w: locked for %v", ErrVerifyCodeMaxTries, SmsVerifyCodeFailedLockTime)
		}
	}

	sendedCodeKey := fmt.Sprintf("sms_verify_code:%s", phoneNumber)
	sendedCode, err := common.RedisGet(sendedCodeKey)

	if err != nil || sendedCode == "" || sendedCode != code {
		// Increment failed counter
		count := 0
		if failedCountStr != "" {
			count, _ = strconv.Atoi(failedCountStr)
		}
		count++

		if count >= MaxSmsVerifyCodeFailedCount {
			common.RedisSet(failedCountKey, strconv.Itoa(MaxSmsVerifyCodeFailedCount), SmsVerifyCodeFailedLockTime)
			return false, fmt.Errorf("%w: locked for %v", ErrVerifyCodeMaxTries, SmsVerifyCodeFailedLockTime)
		}

		common.RedisSet(failedCountKey, strconv.Itoa(count), 0)
		return false, ErrVerifyCodeInvalid
	}

	common.RedisDel(failedCountKey)
	return true, nil
}

func (s *BaseSmsSender) DeleteSms(ctx context.Context, phoneNumber string) error {
	sendedCodeKey := fmt.Sprintf("sms_verify_code:%s", phoneNumber)
	return common.RedisDel(sendedCodeKey)
}

func (s *BaseSmsSender) RandCode(digits int) string {
	if digits <= 0 {
		digits = 6
	}
	// min value: 10^(digits-1)
	// max value: 10^digits - 1
	minVal := int64(1)
	for i := 1; i < digits; i++ {
		minVal *= 10
	}
	maxVal := minVal*10 - 1

	diff := maxVal - minVal + 1
	n, err := rand.Int(rand.Reader, big.NewInt(diff))
	if err != nil {
		// Fallback if crypto/rand fails
		return strconv.FormatInt(minVal, 10)
	}
	result := n.Int64() + minVal
	return strconv.FormatInt(result, 10)
}

// MySmsRequest DTO
type MySmsRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Content  string `json:"content"`
	Mobile   string `json:"mobile"`
	TKey     string `json:"tKey"`
}

// MySmsResponse struct for capturing upstream response
type MySmsResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// MySmsSender represents the custom SMS implementation
type MySmsSender struct {
	BaseSmsSender
}

func NewMySmsSender() *MySmsSender {
	sender := &MySmsSender{}
	sender.Impl = sender // Set the concrete implementation for interface delegation
	return sender
}

func (s *MySmsSender) SendSms(ctx context.Context, phoneNumber string, message string) error {
	tKey := strconv.FormatInt(time.Now().Unix(), 10)

	// MD5 operations: md5(md5(password).hexdigest() + t_key)
	hash1 := md5.Sum([]byte(common.SMSSenderPassword))
	pwdMd5 := hex.EncodeToString(hash1[:])

	hash2 := md5.Sum([]byte(pwdMd5 + tKey))
	finalPwdMd5 := hex.EncodeToString(hash2[:])

	smsReq := MySmsRequest{
		Username: common.SMSSenderUsername,
		Password: finalPwdMd5,
		Content:  message,
		Mobile:   phoneNumber,
		TKey:     tKey,
	}

	// Rule 1: Use common.Marshal instead of encoding/json
	reqBytes, err := common.Marshal(smsReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, common.SMSSenderEndpoint, bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a timeout-controlled context
	ctx, cancel := context.WithTimeout(ctx, DefaultConnectTimeout)
	defer cancel()
	req = req.WithContext(ctx)

	client := GetHttpClient()
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSmsSendFailed, err)
	}
	defer resp.Body.Close()

	var smsResp MySmsResponse
	// Rule 1: Use common.DecodeJson instead of encoding/json
	if err := common.DecodeJson(resp.Body, &smsResp); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if smsResp.Code != http.StatusOK {
		return fmt.Errorf("send failed with code %d: %s", smsResp.Code, smsResp.Msg)
	}

	return nil
}

var smsSenderProvider = map[string]SmsSender{
	"my":      NewMySmsSender(),
	"default": NewMySmsSender(),
}

// GetSmsSender acts as a factory function corresponding to get_sms_sender in Python
func GetSmsSender(name string) SmsSender {
	if sender, ok := smsSenderProvider[name]; ok {
		return sender
	}
	return smsSenderProvider["default"]
}
