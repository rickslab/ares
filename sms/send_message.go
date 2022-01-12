package sms

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rickslab/ares/errcode"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20190711"
	"google.golang.org/grpc/status"
)

var (
	httpClient = &http.Client{
		Timeout: 8 * time.Second,
	}
)

func SendMessageV1(uid string, pwd string, phoneNumber string, msg string) error {
	kvs := url.Values{}
	kvs.Set("uid", uid)
	kvs.Set("pwd", pwd)
	kvs.Set("tos", phoneNumber)
	kvs.Set("msg", msg)
	kvs.Set("otime", "")

	r, err := httpClient.Post("http://service2.winic.org/Service.asmx/SendMessages", "application/x-www-form-urlencoded", bytes.NewReader([]byte(kvs.Encode())))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return status.Errorf(errcode.ErrSmsSendFailed, "sms send-msg-v1 failed: %s", r.Status)
	}
	return nil
}

func SendMessageV2(secretId string, secretKey string, appId string, phoneNumber string, sign string, templateId string, params ...string) error {
	credential := common.NewCredential(
		secretId,
		secretKey,
	)
	cpf := profile.NewClientProfile()
	client, err := sms.NewClient(credential, "ap-guangzhou", cpf)
	if err != nil {
		return err
	}

	request := sms.NewSendSmsRequest()
	request.SmsSdkAppid = &appId
	request.Sign = &sign
	request.TemplateParamSet = common.StringPtrs(params)
	request.TemplateID = &templateId
	request.PhoneNumberSet = common.StringPtrs([]string{fmt.Sprintf("+86%s", phoneNumber)})

	_, err = client.SendSms(request)
	if err != nil {
		if sdkErr, ok := err.(*errors.TencentCloudSDKError); ok {
			return status.Errorf(errcode.ErrSmsSendFailed, "sms send-msg-v2 failed: %+v", *sdkErr)
		}
		return err
	}
	return nil
}
