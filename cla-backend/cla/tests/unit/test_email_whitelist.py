# Copyright The Linux Foundation and each contributor to CommunityBridge.
# SPDX-License-Identifier: MIT

from unittest.mock import patch

import pytest

from cla.models.dynamo_models import Signature, User, UserModel


@pytest.fixture()
def create_user():
    """ Mock user instance """
    with patch.object(User, "__init__", lambda self: None):
        user = User()
        user.model = UserModel()
        yield user


def test_email_against_pattern_with_asterix_prefix(create_user):
    """ Test given user against pattern starting with_asterix_prefix """
    emails = ["harold@bar.com"]
    patterns = ["*bar.com"]
    assert create_user.preprocess_pattern(emails, patterns) == True


def test_subdomain_against_pattern_asterix_prefix(create_user):
    """Test user email on subdomain against pattern """
    emails = ["harold@help.bar.com"]
    patterns = ["*bar.com"]
    assert create_user.preprocess_pattern(emails, patterns) == True


def test_email_multiple_domains(create_user):
    """Test emails against multuple domain lists starting with *.,* and . """
    emails = ["harold@bar.com"]
    patterns = ["*bar.com", "*.bar.com", ".bar.com"]
    assert create_user.preprocess_pattern(emails, patterns) == True
    emails = ["harold@foo.com"]
    assert create_user.preprocess_pattern(emails, patterns) == False


def test_naked_domain(create_user):
    """Test user against naked domain pattern (e.g google.com) """
    emails = ["harold@bar.com"]
    patterns = ["bar.com"]
    assert create_user.preprocess_pattern(emails, patterns) == True
    fail_emails = ["harold@help.bar.com"]
    assert create_user.preprocess_pattern(fail_emails, patterns) == False


def test_pattern_with_asterix_dot_prefix(create_user):
    """ Test given user email against pattern starting with asterix_dot_prefix """
    emails = ["harold@bar.com"]
    patterns = ["*.bar.com"]
    assert create_user.preprocess_pattern(emails, patterns) == True

def test_pattern_with_dot_prefix(create_user):
    """Test given user email against pattern starting with dot_prefix """
    emails = ["harold@bar.com"]
    patterns = [".bar.com"]
    assert create_user.preprocess_pattern(emails, patterns) == True
    domain_emails = ["harold@help.bar.com"]
    assert create_user.preprocess_pattern(domain_emails, patterns) == True
