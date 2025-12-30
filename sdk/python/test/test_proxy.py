# coding: utf-8

# Copyright 2024, Pulumi Corporation.  All rights reserved.

import unittest
import os

from pulumi_esc_sdk.configuration import (
    Configuration,
    _get_proxy_from_environment,
    _get_no_proxy_from_environment,
    _should_bypass_proxy,
)


class TestProxyFunctions(unittest.TestCase):
    """Test proxy-related functions in configuration module"""

    def setUp(self):
        """Clear proxy-related environment variables before each test"""
        self.env_vars_to_clear = [
            'HTTP_PROXY', 'http_proxy',
            'HTTPS_PROXY', 'https_proxy',
            'ALL_PROXY', 'all_proxy',
            'NO_PROXY', 'no_proxy'
        ]
        self.original_env = {}
        for var in self.env_vars_to_clear:
            self.original_env[var] = os.environ.get(var)
            if var in os.environ:
                del os.environ[var]

    def tearDown(self):
        """Restore original environment variables after each test"""
        for var in self.env_vars_to_clear:
            if var in os.environ:
                del os.environ[var]
            if self.original_env[var] is not None:
                os.environ[var] = self.original_env[var]


class TestGetProxyFromEnvironment(TestProxyFunctions):
    """Tests for _get_proxy_from_environment function"""

    def test_https_url_uses_https_proxy_uppercase(self):
        """Test that HTTPS URLs use HTTPS_PROXY environment variable"""
        os.environ['HTTPS_PROXY'] = 'http://proxy.example.com:8080'
        result = _get_proxy_from_environment('https://api.pulumi.com')
        self.assertEqual(result, 'http://proxy.example.com:8080')

    def test_https_url_uses_https_proxy_lowercase(self):
        """Test that HTTPS URLs use https_proxy environment variable"""
        os.environ['https_proxy'] = 'http://proxy.example.com:8080'
        result = _get_proxy_from_environment('https://api.pulumi.com')
        self.assertEqual(result, 'http://proxy.example.com:8080')

    def test_https_proxy_lowercase_takes_precedence(self):
        """Test that https_proxy takes precedence over HTTPS_PROXY"""
        os.environ['HTTPS_PROXY'] = 'http://proxy-upper.example.com:8080'
        os.environ['https_proxy'] = 'http://proxy-lower.example.com:8080'
        result = _get_proxy_from_environment('https://api.pulumi.com')
        self.assertEqual(result, 'http://proxy-lower.example.com:8080')

    def test_http_url_uses_http_proxy_uppercase(self):
        """Test that HTTP URLs use HTTP_PROXY environment variable"""
        os.environ['HTTP_PROXY'] = 'http://proxy.example.com:8080'
        result = _get_proxy_from_environment('http://api.example.com')
        self.assertEqual(result, 'http://proxy.example.com:8080')

    def test_http_url_uses_http_proxy_lowercase(self):
        """Test that HTTP URLs use http_proxy environment variable"""
        os.environ['http_proxy'] = 'http://proxy.example.com:8080'
        result = _get_proxy_from_environment('http://api.example.com')
        self.assertEqual(result, 'http://proxy.example.com:8080')

    def test_http_proxy_lowercase_takes_precedence(self):
        """Test that http_proxy takes precedence over HTTP_PROXY"""
        os.environ['HTTP_PROXY'] = 'http://proxy-upper.example.com:8080'
        os.environ['http_proxy'] = 'http://proxy-lower.example.com:8080'
        result = _get_proxy_from_environment('http://api.example.com')
        self.assertEqual(result, 'http://proxy-lower.example.com:8080')

    def test_fallback_to_all_proxy_uppercase(self):
        """Test fallback to ALL_PROXY when specific proxy not set"""
        os.environ['ALL_PROXY'] = 'http://proxy.example.com:8080'
        result = _get_proxy_from_environment('https://api.pulumi.com')
        self.assertEqual(result, 'http://proxy.example.com:8080')

    def test_fallback_to_all_proxy_lowercase(self):
        """Test fallback to all_proxy when specific proxy not set"""
        os.environ['all_proxy'] = 'http://proxy.example.com:8080'
        result = _get_proxy_from_environment('https://api.pulumi.com')
        self.assertEqual(result, 'http://proxy.example.com:8080')

    def test_all_proxy_lowercase_takes_precedence(self):
        """Test that all_proxy takes precedence over ALL_PROXY"""
        os.environ['ALL_PROXY'] = 'http://proxy-upper.example.com:8080'
        os.environ['all_proxy'] = 'http://proxy-lower.example.com:8080'
        result = _get_proxy_from_environment('https://api.pulumi.com')
        self.assertEqual(result, 'http://proxy-lower.example.com:8080')

    def test_specific_proxy_takes_precedence_over_all_proxy(self):
        """Test that HTTPS_PROXY takes precedence over ALL_PROXY"""
        os.environ['HTTPS_PROXY'] = 'http://https-proxy.example.com:8080'
        os.environ['ALL_PROXY'] = 'http://all-proxy.example.com:8080'
        result = _get_proxy_from_environment('https://api.pulumi.com')
        self.assertEqual(result, 'http://https-proxy.example.com:8080')

    def test_no_proxy_when_none_set(self):
        """Test that None is returned when no proxy is configured"""
        result = _get_proxy_from_environment('https://api.pulumi.com')
        self.assertIsNone(result)

    def test_none_url_returns_none(self):
        """Test that None URL returns None"""
        os.environ['HTTPS_PROXY'] = 'http://proxy.example.com:8080'
        result = _get_proxy_from_environment(None)
        self.assertIsNone(result)

    def test_url_without_scheme_defaults_to_https(self):
        """Test that URLs without scheme default to HTTPS proxy"""
        os.environ['HTTPS_PROXY'] = 'http://https-proxy.example.com:8080'
        os.environ['HTTP_PROXY'] = 'http://http-proxy.example.com:8080'
        result = _get_proxy_from_environment('api.pulumi.com')
        self.assertEqual(result, 'http://https-proxy.example.com:8080')

    def test_case_insensitive_scheme(self):
        """Test that scheme matching is case insensitive"""
        os.environ['HTTPS_PROXY'] = 'http://proxy.example.com:8080'
        result = _get_proxy_from_environment('HTTPS://api.pulumi.com')
        self.assertEqual(result, 'http://proxy.example.com:8080')


class TestGetNoProxyFromEnvironment(TestProxyFunctions):
    """Tests for _get_no_proxy_from_environment function"""

    def test_no_proxy_uppercase(self):
        """Test parsing NO_PROXY environment variable"""
        os.environ['NO_PROXY'] = 'localhost,127.0.0.1,.example.com'
        result = _get_no_proxy_from_environment()
        self.assertEqual(result, ['localhost', '127.0.0.1', '.example.com'])

    def test_no_proxy_lowercase(self):
        """Test parsing no_proxy environment variable"""
        os.environ['no_proxy'] = 'localhost,127.0.0.1,.example.com'
        result = _get_no_proxy_from_environment()
        self.assertEqual(result, ['localhost', '127.0.0.1', '.example.com'])

    def test_no_proxy_lowercase_takes_precedence(self):
        """Test that no_proxy takes precedence over NO_PROXY"""
        os.environ['NO_PROXY'] = 'uppercase.com'
        os.environ['no_proxy'] = 'lowercase.com'
        result = _get_no_proxy_from_environment()
        self.assertEqual(result, ['lowercase.com'])

    def test_empty_no_proxy(self):
        """Test that empty NO_PROXY returns empty list"""
        os.environ['NO_PROXY'] = ''
        result = _get_no_proxy_from_environment()
        self.assertEqual(result, [])

    def test_no_proxy_not_set(self):
        """Test that missing NO_PROXY returns empty list"""
        result = _get_no_proxy_from_environment()
        self.assertEqual(result, [])

    def test_no_proxy_with_spaces(self):
        """Test that spaces around entries are stripped"""
        os.environ['NO_PROXY'] = ' localhost , 127.0.0.1 , .example.com '
        result = _get_no_proxy_from_environment()
        self.assertEqual(result, ['localhost', '127.0.0.1', '.example.com'])

    def test_no_proxy_with_empty_entries(self):
        """Test that empty entries are filtered out"""
        os.environ['NO_PROXY'] = 'localhost,,127.0.0.1,  ,.example.com'
        result = _get_no_proxy_from_environment()
        self.assertEqual(result, ['localhost', '127.0.0.1', '.example.com'])

    def test_no_proxy_single_wildcard(self):
        """Test parsing NO_PROXY with wildcard"""
        os.environ['NO_PROXY'] = '*'
        result = _get_no_proxy_from_environment()
        self.assertEqual(result, ['*'])


class TestShouldBypassProxy(TestProxyFunctions):
    """Tests for _should_bypass_proxy function"""

    def test_exact_match(self):
        """Test exact hostname match"""
        self.assertTrue(_should_bypass_proxy(
            'http://example.com',
            ['example.com']
        ))

    def test_exact_match_case_insensitive(self):
        """Test exact match is case insensitive"""
        self.assertTrue(_should_bypass_proxy(
            'http://Example.COM',
            ['example.com']
        ))

    def test_no_match(self):
        """Test no match returns False"""
        self.assertFalse(_should_bypass_proxy(
            'http://example.com',
            ['different.com']
        ))

    def test_suffix_match_with_leading_dot(self):
        """Test suffix match with leading dot in pattern"""
        self.assertTrue(_should_bypass_proxy(
            'http://api.example.com',
            ['.example.com']
        ))

    def test_suffix_match_without_leading_dot(self):
        """Test suffix match without leading dot in pattern"""
        self.assertTrue(_should_bypass_proxy(
            'http://api.example.com',
            ['example.com']
        ))

    def test_suffix_match_multiple_levels(self):
        """Test suffix match with multiple subdomain levels"""
        self.assertTrue(_should_bypass_proxy(
            'http://deep.nested.api.example.com',
            ['.example.com']
        ))

    def test_suffix_match_respects_domain_boundaries(self):
        """Test that suffix match respects domain boundaries"""
        self.assertFalse(_should_bypass_proxy(
            'http://notexample.com',
            ['example.com']
        ))

    def test_wildcard_matches_all(self):
        """Test that wildcard matches all URLs"""
        self.assertTrue(_should_bypass_proxy(
            'http://any.domain.com',
            ['*']
        ))

    def test_empty_no_proxy_list(self):
        """Test empty no_proxy list returns False"""
        self.assertFalse(_should_bypass_proxy(
            'http://example.com',
            []
        ))

    def test_none_no_proxy_list(self):
        """Test None no_proxy list returns False"""
        self.assertFalse(_should_bypass_proxy(
            'http://example.com',
            None
        ))

    def test_empty_url(self):
        """Test empty URL returns False"""
        self.assertFalse(_should_bypass_proxy(
            '',
            ['example.com']
        ))

    def test_none_url(self):
        """Test None URL returns False"""
        self.assertFalse(_should_bypass_proxy(
            None,
            ['example.com']
        ))

    def test_url_without_hostname(self):
        """Test URL without hostname returns False"""
        self.assertFalse(_should_bypass_proxy(
            'file:///path/to/file',
            ['example.com']
        ))

    def test_localhost_match(self):
        """Test localhost bypass"""
        self.assertTrue(_should_bypass_proxy(
            'http://localhost',
            ['localhost']
        ))

    def test_ip_address_match(self):
        """Test IP address exact match"""
        self.assertTrue(_should_bypass_proxy(
            'http://127.0.0.1',
            ['127.0.0.1']
        ))

    def test_multiple_patterns_first_matches(self):
        """Test multiple patterns where first matches"""
        self.assertTrue(_should_bypass_proxy(
            'http://example.com',
            ['example.com', 'other.com']
        ))

    def test_multiple_patterns_last_matches(self):
        """Test multiple patterns where last matches"""
        self.assertTrue(_should_bypass_proxy(
            'http://example.com',
            ['other.com', 'example.com']
        ))

    def test_multiple_patterns_none_match(self):
        """Test multiple patterns where none match"""
        self.assertFalse(_should_bypass_proxy(
            'http://example.com',
            ['other.com', 'different.com']
        ))

    def test_empty_pattern_ignored(self):
        """Test that empty patterns are ignored"""
        self.assertFalse(_should_bypass_proxy(
            'http://example.com',
            ['', '  ', 'other.com']
        ))

    def test_port_in_url_ignored(self):
        """Test that port in URL doesn't affect matching"""
        self.assertTrue(_should_bypass_proxy(
            'http://example.com:8080',
            ['example.com']
        ))

    def test_port_in_pattern_exact_match(self):
        """Test that pattern with port matches URL with same host and port"""
        self.assertTrue(_should_bypass_proxy(
            'http://example.com:8080',
            ['example.com:8080']
        ))

    def test_port_in_pattern_different_port_no_match(self):
        """Test that pattern with port doesn't match URL with different port"""
        self.assertFalse(_should_bypass_proxy(
            'http://example.com:8080',
            ['example.com:9090']
        ))

    def test_port_in_pattern_no_port_in_url_no_match(self):
        """Test that pattern with port doesn't match URL without port"""
        self.assertFalse(_should_bypass_proxy(
            'http://example.com',
            ['example.com:8080']
        ))

    def test_multiple_ports_in_patterns(self):
        """Test matching against multiple patterns with different ports"""
        self.assertTrue(_should_bypass_proxy(
            'http://example.com:8080',
            ['example.com:9090', 'example.com:8080', 'other.com:8080']
        ))

    def test_pattern_without_port_matches_any_port(self):
        """Test that pattern without port matches URL with any port or no port"""
        patterns = ['example.com']
        self.assertTrue(_should_bypass_proxy('http://example.com', patterns))
        self.assertTrue(_should_bypass_proxy('http://example.com:80', patterns))
        self.assertTrue(_should_bypass_proxy('http://example.com:8080', patterns))
        self.assertTrue(_should_bypass_proxy('http://example.com:443', patterns))

    def test_suffix_match_with_port_in_pattern(self):
        """Test that suffix pattern with port requires matching port"""
        self.assertTrue(_should_bypass_proxy(
            'http://api.example.com:8080',
            ['example.com:8080']
        ))
        self.assertFalse(_should_bypass_proxy(
            'http://api.example.com:9090',
            ['example.com:8080']
        ))

    def test_leading_dot_pattern_with_port(self):
        """Test that leading dot pattern with port matches subdomains with same port"""
        self.assertTrue(_should_bypass_proxy(
            'http://api.example.com:8080',
            ['.example.com:8080']
        ))
        self.assertFalse(_should_bypass_proxy(
            'http://api.example.com:9090',
            ['.example.com:8080']
        ))

    def test_path_in_url_ignored(self):
        """Test that path in URL doesn't affect matching"""
        self.assertTrue(_should_bypass_proxy(
            'http://example.com/path/to/resource',
            ['example.com']
        ))

    def test_leading_dot_pattern_does_not_match_exact_domain(self):
        """Test that pattern '.example.com' does not match exact domain 'example.com'"""
        self.assertFalse(_should_bypass_proxy(
            'http://example.com',
            ['.example.com']
        ))


class TestConfigurationGetProxyForUrl(TestProxyFunctions):
    """Tests for Configuration.get_proxy_for_url method"""

    def test_explicit_proxy_config_takes_precedence(self):
        """Test that explicit proxy configuration takes precedence over environment"""
        os.environ['HTTPS_PROXY'] = 'http://env-proxy.example.com:8080'
        config = Configuration()
        config.proxy = 'http://config-proxy.example.com:8080'

        result = config.get_proxy_for_url('https://api.pulumi.com')
        self.assertEqual(result, 'http://config-proxy.example.com:8080')

    def test_uses_environment_proxy_when_not_configured(self):
        """Test that environment proxy is used when not explicitly configured"""
        os.environ['HTTPS_PROXY'] = 'http://env-proxy.example.com:8080'
        config = Configuration()

        result = config.get_proxy_for_url('https://api.pulumi.com')
        self.assertEqual(result, 'http://env-proxy.example.com:8080')

    def test_returns_none_when_no_proxy_configured(self):
        """Test that None is returned when no proxy is configured"""
        config = Configuration()
        result = config.get_proxy_for_url('https://api.pulumi.com')
        self.assertIsNone(result)

    def test_explicit_no_proxy_bypasses(self):
        """Test that explicit no_proxy configuration bypasses proxy"""
        config = Configuration()
        config.proxy = 'http://proxy.example.com:8080'
        config.no_proxy = ['api.pulumi.com']

        result = config.get_proxy_for_url('https://api.pulumi.com')
        self.assertIsNone(result)

    def test_environment_no_proxy_bypasses(self):
        """Test that environment NO_PROXY bypasses proxy"""
        os.environ['HTTPS_PROXY'] = 'http://proxy.example.com:8080'
        os.environ['NO_PROXY'] = 'api.pulumi.com'
        config = Configuration()

        result = config.get_proxy_for_url('https://api.pulumi.com')
        self.assertIsNone(result)

    def test_explicit_no_proxy_takes_precedence_over_environment(self):
        """Test that explicit no_proxy takes precedence over environment"""
        os.environ['HTTPS_PROXY'] = 'http://proxy.example.com:8080'
        os.environ['NO_PROXY'] = 'env.example.com'
        config = Configuration()
        config.no_proxy = ['config.example.com']

        # Should use explicit no_proxy
        result = config.get_proxy_for_url('https://config.example.com')
        self.assertIsNone(result)

        # Should not use environment no_proxy
        result = config.get_proxy_for_url('https://env.example.com')
        self.assertEqual(result, 'http://proxy.example.com:8080')

    def test_wildcard_no_proxy_bypasses_all(self):
        """Test that wildcard in no_proxy bypasses all URLs"""
        config = Configuration()
        config.proxy = 'http://proxy.example.com:8080'
        config.no_proxy = ['*']

        result = config.get_proxy_for_url('https://any.domain.com')
        self.assertIsNone(result)

    def test_no_proxy_with_multiple_patterns(self):
        """Test no_proxy with multiple patterns"""
        config = Configuration()
        config.proxy = 'http://proxy.example.com:8080'
        config.no_proxy = ['localhost', '127.0.0.1', '.internal.com']

        self.assertIsNone(config.get_proxy_for_url('http://localhost'))
        self.assertIsNone(config.get_proxy_for_url('http://127.0.0.1'))
        self.assertIsNone(config.get_proxy_for_url('http://api.internal.com'))
        self.assertEqual(
            config.get_proxy_for_url('http://external.com'),
            'http://proxy.example.com:8080'
        )

    def test_http_and_https_use_different_proxies(self):
        """Test that HTTP and HTTPS URLs use appropriate proxies"""
        os.environ['HTTPS_PROXY'] = 'http://https-proxy.example.com:8080'
        os.environ['HTTP_PROXY'] = 'http://http-proxy.example.com:8080'
        config = Configuration()

        https_result = config.get_proxy_for_url('https://api.example.com')
        http_result = config.get_proxy_for_url('http://api.example.com')

        self.assertEqual(https_result, 'http://https-proxy.example.com:8080')
        self.assertEqual(http_result, 'http://http-proxy.example.com:8080')

    def test_empty_proxy_string_returns_none(self):
        """Test that empty proxy string is treated as no proxy"""
        os.environ['HTTPS_PROXY'] = ''
        config = Configuration()

        result = config.get_proxy_for_url('https://api.pulumi.com')
        self.assertIsNone(result)

    def test_explicit_empty_proxy_string_returns_none(self):
        """Test that explicitly setting proxy to empty string returns None"""
        os.environ['HTTPS_PROXY'] = 'http://proxy.example.com:8080'
        config = Configuration()
        config.proxy = ''

        result = config.get_proxy_for_url('https://api.pulumi.com')
        self.assertIsNone(result)

    def test_explicit_proxy_with_env_no_proxy(self):
        """Test that explicit proxy respects environment NO_PROXY when config.no_proxy is None"""
        os.environ['NO_PROXY'] = 'internal.com'
        config = Configuration()
        config.proxy = 'http://proxy.example.com:8080'
        # config.no_proxy is None, should use env NO_PROXY

        # Should bypass proxy for internal.com
        result = config.get_proxy_for_url('https://internal.com')
        self.assertIsNone(result)

        # Should use proxy for other domains
        result = config.get_proxy_for_url('https://external.com')
        self.assertEqual(result, 'http://proxy.example.com:8080')

    def test_explicit_proxy_with_empty_no_proxy_list(self):
        """Test that explicit empty no_proxy list does NOT fall back to environment"""
        os.environ['NO_PROXY'] = 'internal.com'
        config = Configuration()
        config.proxy = 'http://proxy.example.com:8080'
        config.no_proxy = []  # Explicitly set to empty list

        # Should NOT use env NO_PROXY, should use proxy for all domains
        result = config.get_proxy_for_url('https://internal.com')
        self.assertEqual(result, 'http://proxy.example.com:8080')

    def test_env_proxy_with_explicit_no_proxy_list(self):
        """Test that environment proxy respects explicit no_proxy list"""
        os.environ['HTTPS_PROXY'] = 'http://proxy.example.com:8080'
        os.environ['NO_PROXY'] = 'env-internal.com'
        config = Configuration()
        config.no_proxy = ['config-internal.com']  # Explicit no_proxy

        # Should bypass based on explicit no_proxy
        result = config.get_proxy_for_url('https://config-internal.com')
        self.assertIsNone(result)

        # Should NOT bypass based on env NO_PROXY
        result = config.get_proxy_for_url('https://env-internal.com')
        self.assertEqual(result, 'http://proxy.example.com:8080')

    def test_both_proxy_and_no_proxy_none(self):
        """Test that both proxy and no_proxy being None uses environment for both"""
        os.environ['HTTPS_PROXY'] = 'http://proxy.example.com:8080'
        os.environ['NO_PROXY'] = 'internal.com'
        config = Configuration()
        # Both config.proxy and config.no_proxy are None

        result = config.get_proxy_for_url('https://internal.com')
        self.assertIsNone(result)

        result = config.get_proxy_for_url('https://external.com')
        self.assertEqual(result, 'http://proxy.example.com:8080')

    def test_explicit_proxy_none_explicit_no_proxy_set(self):
        """Test that explicit proxy=None with explicit no_proxy still uses env proxy"""
        os.environ['HTTPS_PROXY'] = 'http://proxy.example.com:8080'
        config = Configuration()
        config.proxy = None  # Explicitly None
        config.no_proxy = ['internal.com']  # Explicit no_proxy

        # Should use env proxy but explicit no_proxy
        result = config.get_proxy_for_url('https://internal.com')
        self.assertIsNone(result)

        result = config.get_proxy_for_url('https://external.com')
        self.assertEqual(result, 'http://proxy.example.com:8080')

    def test_explicit_proxy_with_explicit_no_proxy(self):
        """Test that both explicit proxy and no_proxy work together"""
        config = Configuration()
        config.proxy = 'http://proxy.example.com:8080'
        config.no_proxy = ['localhost', '127.0.0.1', '.internal.com']

        # Should bypass for patterns in no_proxy
        self.assertIsNone(config.get_proxy_for_url('http://localhost'))
        self.assertIsNone(config.get_proxy_for_url('http://127.0.0.1'))
        self.assertIsNone(config.get_proxy_for_url('http://api.internal.com'))

        # Should use proxy for other domains
        result = config.get_proxy_for_url('https://external.com')
        self.assertEqual(result, 'http://proxy.example.com:8080')

    def test_no_proxy_none_vs_empty_list_behavior(self):
        """Test the difference between no_proxy=None and no_proxy=[]"""
        os.environ['NO_PROXY'] = 'env-bypass.com'
        config = Configuration()
        config.proxy = 'http://proxy.example.com:8080'

        # When no_proxy is None, should use environment NO_PROXY
        config.no_proxy = None
        result = config.get_proxy_for_url('https://env-bypass.com')
        self.assertIsNone(result)

        # When no_proxy is empty list, should NOT use environment NO_PROXY
        config.no_proxy = []
        result = config.get_proxy_for_url('https://env-bypass.com')
        self.assertEqual(result, 'http://proxy.example.com:8080')

    def test_proxy_and_no_proxy_independent_behavior(self):
        """Test that proxy and no_proxy settings are independent"""
        os.environ['HTTPS_PROXY'] = 'http://env-proxy.example.com:8080'
        os.environ['NO_PROXY'] = 'env-bypass.com'
        config = Configuration()

        # Explicit proxy, environment no_proxy
        config.proxy = 'http://explicit-proxy.example.com:8080'
        config.no_proxy = None
        result = config.get_proxy_for_url('https://env-bypass.com')
        self.assertIsNone(result)  # Bypassed by env NO_PROXY
        result = config.get_proxy_for_url('https://external.com')
        self.assertEqual(result, 'http://explicit-proxy.example.com:8080')  # Explicit proxy

        # Environment proxy, explicit no_proxy
        config.proxy = None
        config.no_proxy = ['explicit-bypass.com']
        result = config.get_proxy_for_url('https://explicit-bypass.com')
        self.assertIsNone(result)  # Bypassed by explicit no_proxy
        result = config.get_proxy_for_url('https://env-bypass.com')
        self.assertEqual(result, 'http://env-proxy.example.com:8080')  # Not bypassed


if __name__ == '__main__':
    unittest.main()
