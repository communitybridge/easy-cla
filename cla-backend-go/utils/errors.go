// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package utils

import (
	"fmt"
	"strings"
)

// ConversionError is an error model for representing conversion errors
type ConversionError struct {
	FromType string
	ToType   string
	Err      error
}

// Error is an error string function for Salesforce Project not found errors
func (e *ConversionError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("unable to convert %s to %s", e.FromType, e.ToType)
	}
	return fmt.Sprintf("unable to convert %s to %s due to error: %+v", e.FromType, e.ToType, e.Err)
}

// Unwrap method returns its contained error
func (e *ConversionError) Unwrap() error {
	return e.Err
}

// SFProjectNotFound is an error model for Salesforce Project not found errors
type SFProjectNotFound struct {
	ProjectSFID string
	Err         error
}

// Error is an error string function for Salesforce Project not found errors
func (e *SFProjectNotFound) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("salesforce project %s not found", e.ProjectSFID)
	}
	return fmt.Sprintf("salesforce project %s not found: %+v", e.ProjectSFID, e.Err)
}

// Unwrap method returns its contained error
func (e *SFProjectNotFound) Unwrap() error {
	return e.Err
}

// CLAGroupNotFound is an error model for CLA Group not found errors
type CLAGroupNotFound struct {
	CLAGroupID string
	Err        error
}

// Error is an error string function for CLA Group not found errors
func (e *CLAGroupNotFound) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("cla group %s not found", e.CLAGroupID)
	}
	return fmt.Sprintf("cla group %s not found: %+v", e.CLAGroupID, e.Err)
}

// Unwrap method returns its contained error
func (e *CLAGroupNotFound) Unwrap() error {
	return e.Err
}

// CLAGroupNameConflict is an error model for CLA Group name conflicts
type CLAGroupNameConflict struct {
	CLAGroupID   string
	CLAGroupName string
	Err          error
}

// Error is an error string function for CLA Group not found errors
func (e *CLAGroupNameConflict) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("cla group ID: %s, name: %s, conflict", e.CLAGroupID, e.CLAGroupName)
	}
	return fmt.Sprintf("cla group ID: %s, name: %s, conflict, error: %+v", e.CLAGroupID, e.CLAGroupName, e.Err)
}

// Unwrap method returns its contained error
func (e *CLAGroupNameConflict) Unwrap() error {
	return e.Err
}

// CLAGroupICLANotConfigured is an error model for CLA Group ICLA not configured
type CLAGroupICLANotConfigured struct {
	CLAGroupID   string
	CLAGroupName string
	Err          error
}

// Error is an error string function for CLA Group ICLA not configured
func (e *CLAGroupICLANotConfigured) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("cla group %s (%s) is not configured for ICLAs", e.CLAGroupName, e.CLAGroupID)
	}
	return fmt.Sprintf("cla group %s (%s) is not configured for ICLAs: %+v", e.CLAGroupName, e.CLAGroupID, e.Err)
}

// Unwrap method returns its contained error
func (e *CLAGroupICLANotConfigured) Unwrap() error {
	return e.Err
}

// CLAGroupCCLANotConfigured is an error model for CLA Group CCLA not configured
type CLAGroupCCLANotConfigured struct {
	CLAGroupID   string
	CLAGroupName string
	Err          error
}

// Error is an error string function for CLA Group CCLA not configured
func (e *CLAGroupCCLANotConfigured) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("cla group %s (%s) is not configured for CCLAs", e.CLAGroupName, e.CLAGroupID)
	}
	return fmt.Sprintf("cla group %s (%s) is not configured for CCLAs: %+v", e.CLAGroupName, e.CLAGroupID, e.Err)
}

// Unwrap method returns its contained error
func (e *CLAGroupCCLANotConfigured) Unwrap() error {
	return e.Err
}

// ProjectCLAGroupMappingNotFound is an error model for project CLA Group not found errors
type ProjectCLAGroupMappingNotFound struct {
	ProjectSFID string
	CLAGroupID  string
	Err         error
}

// Error is an error string function for project CLA Group not found errors
func (e *ProjectCLAGroupMappingNotFound) Error() string {
	if e.CLAGroupID == "" {
		if e.Err == nil {
			return fmt.Sprintf("project %s to cla group mapping not found", e.ProjectSFID)
		}
		return fmt.Sprintf("project %s cla group mapping not found: %+v", e.ProjectSFID, e.Err)
	}
	if e.ProjectSFID == "" {
		if e.Err == nil {
			return fmt.Sprintf("project to cla group %s mapping not found", e.CLAGroupID)
		}
		return fmt.Sprintf("project cla group %s mapping not found: %+v", e.CLAGroupID, e.Err)
	}

	return fmt.Sprintf("project %s cla group %s mapping not found: %+v", e.ProjectSFID, e.CLAGroupID, e.Err)
}

// Unwrap method returns its contained error
func (e *ProjectCLAGroupMappingNotFound) Unwrap() error {
	return e.Err
}

// GitHubOrgNotFound is an error model for GitHub Organization not found errors
type GitHubOrgNotFound struct {
	ProjectSFID      string
	OrganizationName string
	Err              error
}

// Error is an error string function for project CLA Group not found errors
func (e *GitHubOrgNotFound) Error() string {
	return fmt.Sprintf("github organization with name: %s and projectSFID: %s not found: %+v", e.OrganizationName, e.ProjectSFID, e.Err)
}

// Unwrap method returns its contained error
func (e *GitHubOrgNotFound) Unwrap() error {
	return e.Err
}

// CompanyAdminNotFound is an error model for Salesforce Project not found errors
type CompanyAdminNotFound struct {
	CompanySFID string
	Err         error
}

// Error is an error string function for Salesforce Project not found errors
func (e *CompanyAdminNotFound) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("company admin for company with ID %s not found", e.CompanySFID)
	}
	return fmt.Sprintf("company admin for company with ID %s not found: %+v", e.CompanySFID, e.Err)
}

// Unwrap method returns its contained error
func (e *CompanyAdminNotFound) Unwrap() error {
	return e.Err
}

// CompanyNotFound is an error model for company not found errors
type CompanyNotFound struct {
	Message                  string
	CompanyID                string
	CompanySFID              string
	CompanyName              string
	CompanySigningEntityName string
	Err                      error
}

// Error is an error string function for Salesforce Project not found errors
func (e *CompanyNotFound) Error() string {
	msg := "company does not exist "
	if e.Message != "" {
		msg = e.Message
	}
	if e.CompanyName != "" {
		msg = fmt.Sprintf("%s - company name: %s ", msg, e.CompanyName)
	}
	if e.CompanySigningEntityName != "" {
		msg = fmt.Sprintf("%s - company sigining entity name: %s ", msg, e.CompanySigningEntityName)
	}
	if e.CompanyID != "" {
		msg = fmt.Sprintf("%s - company ID: %s ", msg, e.CompanyID)
	}
	if e.CompanySFID != "" {
		msg = fmt.Sprintf("%s - company SFID: %s ", msg, e.CompanySFID)
	}
	if e.Err != nil {
		msg = fmt.Sprintf("%s - error: %+v ", msg, e.Err.Error())
	}

	return strings.TrimSpace(msg)
}

// Unwrap method returns its contained error
func (e *CompanyNotFound) Unwrap() error {
	return e.Err
}
