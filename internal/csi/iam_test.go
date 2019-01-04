package csi

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
)

// TODO: create tests for createIAM()
// TODO: create tests for getManualDeleteCommands()
// TODO: create tests for deleteIAMResources() when invalid iamResources

func TestIAM(t *testing.T) {
	if os.Getenv("RUN_AWS_TESTS") != "1" {
		t.Skip()
	}

	// TESTS CREATE
	resource, err := createIAMResources("us-west-2")
	if err != nil {
		t.Fatal(err)
	}

	if resource == nil {
		t.Fatal("expected 'resources' not to be nil")
	}
	if resource.svc == nil {
		t.Fatalf("expected 'svc' not to be nil")
	}
	if resource.lg == nil {
		t.Fatalf("expected 'lg' not to be nil")
	}
	if resource.instanceProfile == nil {
		t.Fatalf("expected 'instanceProfile' not to be nil")
	}
	if resource.policy == nil {
		t.Fatalf("expected 'policy' not to be nil")
	}
	if resource.role == nil {
		t.Fatalf("expected 'role' not to be nil")
	}

	// Check if instance profile exists
	instanceProfileOutput, err := resource.svc.GetInstanceProfile(&iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(resource.instanceProfile.name),
	})
	if err != nil {
		t.Fatalf("expected no error when getting instance profile, got %v", err)
	}
	instanceProfile := instanceProfileOutput.InstanceProfile
	if resource.instanceProfile.arn != *instanceProfile.Arn {
		t.Fatalf("expected %q for instance profile arn, got %q",
			resource.instanceProfile.arn,
			*instanceProfile.Arn,
		)
	}

	// Check if role exists
	roleOutput, err := resource.svc.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(resource.role.name),
	})
	if err != nil {
		t.Fatalf("expected no error when getting role, got %v", err)
	}
	role := roleOutput.Role
	if resource.role.arn != *role.Arn {
		t.Fatalf("expected %q for role arn, got %q", resource.role.arn, *role.Arn)
	}

	// Checks if policy exists
	policyOutput, err := resource.svc.GetPolicy(&iam.GetPolicyInput{
		PolicyArn: aws.String(resource.policy.arn),
	})
	if err != nil {
		t.Fatalf("expected no error when getting policy, got %v", err)
	}
	policy := policyOutput.Policy
	if resource.policy.name != *policy.PolicyName{
		t.Fatalf("expected %q for policy name, got %q", resource.policy.name, *policy.PolicyName)
	}

	// Checks if role is attached to instance profile
	if len(instanceProfile.Roles) != 1 {
		t.Fatalf("expected instance profile %q to have one role, got %v",
			resource.instanceProfile.name,
			len(instanceProfile.Roles),
		)
	}
	if resource.role.name != *instanceProfile.Roles[0].RoleName {
		t.Fatalf("expected instance profile %q to have role %q, got %q",
			resource.instanceProfile.name,
			resource.role.name,
			*instanceProfile.Roles[0].RoleName,
		)
	}

	// Checks if policy is attached to role
	attachedRolePolicyOutput, err := resource.svc.ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(resource.role.name),
	})
	if err != nil {
		t.Fatalf("expected no error when getting list of attached role/policies, got %v", err)
	}
	if len(attachedRolePolicyOutput.AttachedPolicies) != 1 {
		t.Fatalf("expected role %q to have one policy, got %v",
			resource.role.name,
			len(attachedRolePolicyOutput.AttachedPolicies),
		)
	}
	if *attachedRolePolicyOutput.AttachedPolicies[0].PolicyName != resource.policy.name {
		t.Fatalf("expected role %q to have policy %q, got %v",
			resource.role.name,
			resource.policy.name,
			*attachedRolePolicyOutput.AttachedPolicies[0].PolicyName,
		)
	}

	deletedInstanceProfileName := resource.instanceProfile.name
	deletedRoleName := resource.role.name
	deletedPolicyArn := resource.policy.arn

	// TESTS DELETE
	err = resource.deleteIAMResources()
	if err != nil {
		t.Fatalf("expected no error when deleteing resource, got %v", err)
	}

	if resource == nil {
		t.Fatal("expected 'resources' not to be nil")
	}
	if resource.svc == nil {
		t.Fatalf("expected 'svc' not to be nil")
	}
	if resource.lg == nil {
		t.Fatalf("expected 'lg' not to be nil")
	}
	if resource.instanceProfile != nil {
		t.Fatalf("expected 'instanceProfile' to be nil, got %v", resource.instanceProfile)
	}
	if resource.policy != nil {
		t.Fatalf("expected 'policy' to be nil, got %v", resource.policy)
	}
	if resource.role != nil {
		t.Fatalf("expected 'role' to be nil, got %v", resource.role)
	}

	// Check if instance profile has been deleted
	instanceProfileOutputAfterDelete, err := resource.svc.GetInstanceProfile(&iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(deletedInstanceProfileName),
	})
	if err == nil {
		t.Fatalf("expected error when getting instance profile %q after delete", deletedInstanceProfileName)
	}
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() != iam.ErrCodeNoSuchEntityException {
			t.Fatalf("expected error %q when getting instance profile %q after delete, got %q",
				iam.ErrCodeNoSuchEntityException,
				deletedInstanceProfileName,
				awsErr.Code(),
			)
		}
	}
	if *instanceProfileOutputAfterDelete != (iam.GetInstanceProfileOutput{}) {
		t.Fatalf("expected deleted instance profile output to be empty, got %#v", instanceProfileOutputAfterDelete)
	}

	// Check if role has been deleted
	roleOutputAfterDelete, err := resource.svc.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(deletedRoleName),
	})
	if err == nil {
		t.Fatalf("expected error when getting role %q after delete", deletedRoleName)
	}
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() != iam.ErrCodeNoSuchEntityException {
			t.Fatalf("expected error %q when getting role %q after delete, got %q",
				iam.ErrCodeNoSuchEntityException,
				deletedRoleName,
				awsErr.Code())
		}
	}
	if *roleOutputAfterDelete != (iam.GetRoleOutput{}) {
		t.Fatalf("expected deleted role output to be empty, got %#v", roleOutputAfterDelete)
	}

	// Checks if policy has been deleted
	policyOutputAfterDelete, err := resource.svc.GetPolicy(&iam.GetPolicyInput{
		PolicyArn: aws.String(deletedPolicyArn),
	})
	if err == nil {
		t.Fatalf("expected error when getting policy %q (arn) after delete", deletedPolicyArn)
	}
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() != iam.ErrCodeNoSuchEntityException {
			t.Fatalf("expected error %q when getting policy %q (arn) after delete, got %q",
				iam.ErrCodeNoSuchEntityException,
				deletedPolicyArn,
				awsErr.Code(),
			)
		}
	}
	if *policyOutputAfterDelete != (iam.GetPolicyOutput{}) {
		t.Fatalf("expected deleted policy output to be empty, got %#v", policyOutputAfterDelete)
	}

}

func TestGetManualDeleteCommands(t *testing.T) {
	//if os.Getenv("RUN_AWS_TESTS") != "1" {
	//	t.Skip()
	//}

	// Empty iamResources{}
	testIamResources := iamResources{}

	actualResults := testIamResources.getManualDeleteCommands()
	expectedResults := ""

	if actualResults != expectedResults {
		t.Fatalf("expected delete commands when empty: %s\n, got: %s", expectedResults, actualResults)
	}

	// With only instanceProfile
	testIamResources.instanceProfile = &iamResource {
		name: "instance-profile-test-name",
		arn: "instance-profile-test-arn",
	}

	expectedDeleteInstanceProfile := fmt.Sprintf("aws iam delete-instance-profile --instance-profile-name %s", testIamResources.instanceProfile.name)
	actualResults = testIamResources.getManualDeleteCommands()

	if actualResults != expectedDeleteInstanceProfile {
		t.Fatalf("expected delete commands when only instance profile: %q,\n got: %q", expectedDeleteInstanceProfile, actualResults)
	}
}

//expectedResults := `aws iam remove-role-from-instance-profile --instance-profile-name instance-profile-test-name --role-name role-test-name \
// aws iam delete-instance-profile --instance-profile-name instance-profile-test-name \
// aws iam detach-role-policy --role-name role-test-name --policy-arn policy-test-arn \
// aws iam delete-role --role-name role-test-name \
// aws iam delete-policy --policy-arn policy-test-arn`

//testIamResources.instanceProfile := iamResources {
//	instanceProfile: &iamResource {
//		name: "instance-profile-test-name",
//		arn: "instance-profile-test-arn",
//	},
//	policy: &iamResource {
//		name: "policy-test-name",
//		arn: "policy-test-arn",
//	},
//	role: &iamResource {
//		name: "role-test-name",
//		arn: "role-test-arn",
//	},
//}
