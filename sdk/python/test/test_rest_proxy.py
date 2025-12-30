# coding: utf-8

# Copyright 2024, Pulumi Corporation.  All rights reserved.

import unittest
import os
from unittest.mock import Mock, patch

from pulumi_esc_sdk.configuration import Configuration
from pulumi_esc_sdk.rest import RESTClientObject


class TestRESTClientProxyBehavior(unittest.TestCase):
    """Test that RESTClientObject properly uses proxy settings"""

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

    @patch('pulumi_esc_sdk.rest.urllib3.ProxyManager')
    @patch('pulumi_esc_sdk.rest.urllib3.PoolManager')
    def test_no_proxy_uses_pool_manager(self, mock_pool_manager, mock_proxy_manager):
        """Test that requests without proxy use PoolManager"""
        config = Configuration()
        rest_client = RESTClientObject(config)

        # Mock the pool manager's request method
        mock_pool_instance = Mock()
        mock_pool_instance.request.return_value = Mock(status=200, data=b'{}', headers={})
        mock_pool_manager.return_value = mock_pool_instance

        # Make a request
        rest_client.request('GET', 'https://api.pulumi.com/test')

        # Verify PoolManager was created (not ProxyManager)
        mock_pool_manager.assert_called_once()
        mock_proxy_manager.assert_not_called()

        # Verify request was made
        mock_pool_instance.request.assert_called_once()

    @patch('pulumi_esc_sdk.rest.urllib3.ProxyManager')
    @patch('pulumi_esc_sdk.rest.urllib3.PoolManager')
    def test_explicit_proxy_uses_proxy_manager(self, mock_pool_manager, mock_proxy_manager):
        """Test that explicit proxy configuration uses ProxyManager"""
        config = Configuration()
        config.proxy = 'http://proxy.example.com:8080'
        rest_client = RESTClientObject(config)

        # Mock the proxy manager's request method
        mock_proxy_instance = Mock()
        mock_proxy_instance.request.return_value = Mock(status=200, data=b'{}', headers={})
        mock_proxy_manager.return_value = mock_proxy_instance

        # Make a request
        rest_client.request('GET', 'https://api.pulumi.com/test')

        # Verify ProxyManager was created with correct proxy URL
        mock_proxy_manager.assert_called_once()
        call_kwargs = mock_proxy_manager.call_args[1]
        self.assertEqual(call_kwargs['proxy_url'], 'http://proxy.example.com:8080')

        # Verify PoolManager was not used
        mock_pool_manager.assert_not_called()

        # Verify request was made
        mock_proxy_instance.request.assert_called_once()

    @patch('pulumi_esc_sdk.rest.urllib3.ProxyManager')
    @patch('pulumi_esc_sdk.rest.urllib3.PoolManager')
    def test_different_urls_can_use_different_managers(self, mock_pool_manager, mock_proxy_manager):
        """Test that different URLs can use different pool managers based on no_proxy"""
        config = Configuration()
        config.proxy = 'http://proxy.example.com:8080'
        config.no_proxy = ['internal.com']
        rest_client = RESTClientObject(config)

        # Mock both managers
        mock_pool_instance = Mock()
        mock_pool_instance.request.return_value = Mock(status=200, data=b'{}', headers={})
        mock_pool_manager.return_value = mock_pool_instance

        mock_proxy_instance = Mock()
        mock_proxy_instance.request.return_value = Mock(status=200, data=b'{}', headers={})
        mock_proxy_manager.return_value = mock_proxy_instance

        # Request to internal domain (should bypass proxy)
        rest_client.request('GET', 'https://internal.com/api')
        self.assertEqual(mock_pool_manager.call_count, 1)
        self.assertEqual(mock_proxy_manager.call_count, 0)

        # Request to external domain (should use proxy)
        rest_client.request('GET', 'https://external.com/api')
        self.assertEqual(mock_pool_manager.call_count, 1)
        self.assertEqual(mock_proxy_manager.call_count, 1)

        # Verify both managers had requests made
        mock_pool_instance.request.assert_called_once()
        mock_proxy_instance.request.assert_called_once()

    @patch('pulumi_esc_sdk.rest.urllib3.ProxyManager')
    @patch('pulumi_esc_sdk.rest.urllib3.PoolManager')
    def test_pool_manager_caching(self, mock_pool_manager, mock_proxy_manager):
        """Test that pool managers are cached and reused"""
        config = Configuration()
        config.proxy = 'http://proxy.example.com:8080'
        rest_client = RESTClientObject(config)

        # Mock the proxy manager's request method
        mock_proxy_instance = Mock()
        mock_proxy_instance.request.return_value = Mock(status=200, data=b'{}', headers={})
        mock_proxy_manager.return_value = mock_proxy_instance

        # Make multiple requests to same domain
        rest_client.request('GET', 'https://api.pulumi.com/test1')
        rest_client.request('GET', 'https://api.pulumi.com/test2')
        rest_client.request('GET', 'https://api.pulumi.com/test3')

        # Verify ProxyManager was created only once (cached)
        mock_proxy_manager.assert_called_once()

        # Verify all three requests were made on the same instance
        self.assertEqual(mock_proxy_instance.request.call_count, 3)

    @patch('pulumi_esc_sdk.rest.urllib3.ProxyManager')
    @patch('pulumi_esc_sdk.rest.urllib3.PoolManager')
    def test_proxy_headers_passed_to_proxy_manager(self, mock_pool_manager, mock_proxy_manager):
        """Test that proxy_headers are passed to ProxyManager"""
        config = Configuration()
        config.proxy = 'http://proxy.example.com:8080'
        config.proxy_headers = {'Proxy-Authorization': 'Bearer token123'}
        rest_client = RESTClientObject(config)

        # Mock the proxy manager
        mock_proxy_instance = Mock()
        mock_proxy_instance.request.return_value = Mock(status=200, data=b'{}', headers={})
        mock_proxy_manager.return_value = mock_proxy_instance

        # Make a request
        rest_client.request('GET', 'https://api.pulumi.com/test')

        # Verify ProxyManager was created with proxy_headers
        mock_proxy_manager.assert_called_once()
        call_kwargs = mock_proxy_manager.call_args[1]
        self.assertEqual(call_kwargs['proxy_headers'], {'Proxy-Authorization': 'Bearer token123'})

    @patch('pulumi_esc_sdk.rest.urllib3.PoolManager')
    def test_socks_proxy_uses_socks_manager(self, mock_pool_manager):
        """Test that SOCKS proxy URL uses SOCKSProxyManager"""
        # Mock the SOCKSProxyManager import and class
        mock_socks_manager_class = Mock()
        mock_socks_instance = Mock()
        mock_socks_instance.request.return_value = Mock(status=200, data=b'{}', headers={})
        mock_socks_manager_class.return_value = mock_socks_instance

        with patch.dict('sys.modules', {'urllib3.contrib.socks': Mock(SOCKSProxyManager=mock_socks_manager_class)}):
            config = Configuration()
            config.proxy = 'socks5://proxy.example.com:1080'
            rest_client = RESTClientObject(config)

            # Make a request
            rest_client.request('GET', 'https://api.pulumi.com/test')

            # Verify SOCKSProxyManager was created
            mock_socks_manager_class.assert_called_once()
            call_kwargs = mock_socks_manager_class.call_args[1]
            self.assertEqual(call_kwargs['proxy_url'], 'socks5://proxy.example.com:1080')

            # Verify request was made
            mock_socks_instance.request.assert_called_once()

    @patch('pulumi_esc_sdk.rest.urllib3.ProxyManager')
    @patch('pulumi_esc_sdk.rest.urllib3.PoolManager')
    def test_ssl_settings_passed_to_managers(self, mock_pool_manager, mock_proxy_manager):
        """Test that SSL settings from configuration are passed to pool managers"""
        config = Configuration()
        config.verify_ssl = False
        config.ssl_ca_cert = '/path/to/ca.pem'
        rest_client = RESTClientObject(config)

        # Mock the pool manager
        mock_pool_instance = Mock()
        mock_pool_instance.request.return_value = Mock(status=200, data=b'{}', headers={})
        mock_pool_manager.return_value = mock_pool_instance

        # Make a request without proxy
        rest_client.request('GET', 'https://api.pulumi.com/test')

        # Verify PoolManager was created with SSL settings
        mock_pool_manager.assert_called_once()
        call_kwargs = mock_pool_manager.call_args[1]
        self.assertIn('cert_reqs', call_kwargs)
        self.assertEqual(call_kwargs['ca_certs'], '/path/to/ca.pem')


if __name__ == '__main__':
    unittest.main()
