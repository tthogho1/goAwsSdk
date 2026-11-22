package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

// IAM role listing. Output is one line per role:
//
//	Role: <name>, <created YYYY-MM-DD>, <path>, <arn>, <description>
//
// The description is last because it may contain commas.

// ListIAMRoles prints all IAM roles whose name matches pattern (a regular
// expression); an empty pattern lists every role.
func ListIAMRoles(svc *iam.IAM, pattern string) error {
	var re *regexp.Regexp
	if pattern != "" {
		var err error
		if re, err = regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid pattern %q: %w", pattern, err)
		}
	}

	count := 0
	err := svc.ListRolesPages(&iam.ListRolesInput{}, func(page *iam.ListRolesOutput, _ bool) bool {
		for _, r := range page.Roles {
			name := aws.StringValue(r.RoleName)
			if re != nil && !re.MatchString(name) {
				continue
			}
			created := ""
			if r.CreateDate != nil {
				created = r.CreateDate.Format("2006-01-02")
			}
			fmt.Printf("Role: %s, %s, %s, %s, %s\n",
				name, created, aws.StringValue(r.Path), aws.StringValue(r.Arn), aws.StringValue(r.Description))
			count++
		}
		return true
	})
	if err != nil {
		return err
	}
	fmt.Printf("Total: %d\n", count)
	return nil
}

// roleDetail is the JSON document printed by DescribeIAMRole.
type roleDetail struct {
	Role             *iam.Role
	TrustPolicy      json.RawMessage
	AttachedPolicies []attachedPolicy
	InlinePolicies   []inlinePolicy
}

type attachedPolicy struct {
	PolicyName string
	PolicyArn  string
	VersionId  string
	Document   json.RawMessage
}

type inlinePolicy struct {
	PolicyName string
	Document   json.RawMessage
}

// DescribeIAMRole writes roleName's details, trust policy and all attached and
// inline policy documents as one indented JSON object to w.
func DescribeIAMRole(svc *iam.IAM, roleName string, w io.Writer) error {
	if roleName == "" {
		return fmt.Errorf("role name is required (-n <role>)")
	}

	roleOut, err := svc.GetRole(&iam.GetRoleInput{RoleName: aws.String(roleName)})
	if err != nil {
		return err
	}
	role := roleOut.Role
	detail := roleDetail{Role: role, AttachedPolicies: []attachedPolicy{}, InlinePolicies: []inlinePolicy{}}
	if detail.TrustPolicy, err = decodePolicyDocument(aws.StringValue(role.AssumeRolePolicyDocument)); err != nil {
		return fmt.Errorf("trust policy: %w", err)
	}
	role.AssumeRolePolicyDocument = nil // shown decoded as TrustPolicy

	var arns []*string
	err = svc.ListAttachedRolePoliciesPages(&iam.ListAttachedRolePoliciesInput{RoleName: aws.String(roleName)},
		func(page *iam.ListAttachedRolePoliciesOutput, _ bool) bool {
			for _, p := range page.AttachedPolicies {
				arns = append(arns, p.PolicyArn)
			}
			return true
		})
	if err != nil {
		return err
	}
	for _, arn := range arns {
		pol, err := svc.GetPolicy(&iam.GetPolicyInput{PolicyArn: arn})
		if err != nil {
			return err
		}
		ver, err := svc.GetPolicyVersion(&iam.GetPolicyVersionInput{PolicyArn: arn, VersionId: pol.Policy.DefaultVersionId})
		if err != nil {
			return err
		}
		doc, err := decodePolicyDocument(aws.StringValue(ver.PolicyVersion.Document))
		if err != nil {
			return fmt.Errorf("policy %s: %w", aws.StringValue(arn), err)
		}
		detail.AttachedPolicies = append(detail.AttachedPolicies, attachedPolicy{
			PolicyName: aws.StringValue(pol.Policy.PolicyName),
			PolicyArn:  aws.StringValue(arn),
			VersionId:  aws.StringValue(pol.Policy.DefaultVersionId),
			Document:   doc,
		})
	}

	var names []*string
	err = svc.ListRolePoliciesPages(&iam.ListRolePoliciesInput{RoleName: aws.String(roleName)},
		func(page *iam.ListRolePoliciesOutput, _ bool) bool {
			names = append(names, page.PolicyNames...)
			return true
		})
	if err != nil {
		return err
	}
	for _, name := range names {
		out, err := svc.GetRolePolicy(&iam.GetRolePolicyInput{RoleName: aws.String(roleName), PolicyName: name})
		if err != nil {
			return err
		}
		doc, err := decodePolicyDocument(aws.StringValue(out.PolicyDocument))
		if err != nil {
			return fmt.Errorf("inline policy %s: %w", aws.StringValue(name), err)
		}
		detail.InlinePolicies = append(detail.InlinePolicies, inlinePolicy{PolicyName: aws.StringValue(name), Document: doc})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(detail)
}

// decodePolicyDocument turns the URL-encoded policy JSON returned by IAM into
// raw JSON so it is embedded as an object, not a string.
func decodePolicyDocument(encoded string) (json.RawMessage, error) {
	if encoded == "" {
		return json.RawMessage("null"), nil
	}
	decoded, err := url.QueryUnescape(encoded)
	if err != nil {
		return nil, err
	}
	if !json.Valid([]byte(decoded)) {
		return nil, fmt.Errorf("policy document is not valid JSON")
	}
	return json.RawMessage(decoded), nil
}
