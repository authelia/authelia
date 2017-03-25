
var assert = require('assert');
var winston = require('winston');
var AccessControl = require('../../src/lib/access_control');

describe('test access control manager', function() {
  var access_control;
  var acl_config;
  var acl_builder;
  var acl_matcher;

  beforeEach(function() {
    acl_config = {};
    access_control = AccessControl(winston, acl_config);
    acl_builder = access_control.builder;
    acl_matcher = access_control.matcher;
  });

  describe('building user group access control matcher', function() {
    it('should deny all if nothing is defined in the config', function() {
      var allowed_domains = acl_builder.get_allowed_domains('user', ['group1', 'group2']); 
      assert.deepEqual(allowed_domains, []);
    });

    it('should allow domain test.example.com to all users if defined in' + 
       ' default policy', function() {
      acl_config.default = ['test.example.com'];
      
      var allowed_domains = acl_builder.get_allowed_domains('user', ['group1', 'group2']); 
      assert.deepEqual(allowed_domains, ['test.example.com']);
    });

    it('should allow domain test.example.com to all users in group mygroup', function() {
      var allowed_domains0 = acl_builder.get_allowed_domains('user', ['group1', 'group1']); 
      assert.deepEqual(allowed_domains0, []);

      acl_config.groups = {
        mygroup: ['test.example.com']
      };

      var allowed_domains1 = acl_builder.get_allowed_domains('user', ['group1', 'group2']); 
      assert.deepEqual(allowed_domains1, []);

      var allowed_domains2 = acl_builder.get_allowed_domains('user', ['group1', 'mygroup']); 
      assert.deepEqual(allowed_domains2, ['test.example.com']);
    });

    it('should allow domain test.example.com based on per user config', function() {
      var allowed_domains0 = acl_builder.get_allowed_domains('user', ['group1']); 
      assert.deepEqual(allowed_domains0, []);

      acl_config.users = {
        user1: ['test.example.com']
      };

      var allowed_domains1 = acl_builder.get_allowed_domains('user', ['group1', 'mygroup']); 
      assert.deepEqual(allowed_domains1, []);

      var allowed_domains2 = acl_builder.get_allowed_domains('user1', ['group1', 'mygroup']); 
      assert.deepEqual(allowed_domains2, ['test.example.com']);
    });

    it('should allow domains from user and groups', function() {
      acl_config.groups = {
        group2: ['secret.example.com', 'secret1.example.com']
      };
      acl_config.users = {
        user: ['test.example.com']
      };

      var allowed_domains0 = acl_builder.get_allowed_domains('user', ['group1', 'group2']); 
      assert.deepEqual(allowed_domains0, [
        'secret.example.com',
        'secret1.example.com',
        'test.example.com', 
      ]);
    });

    it('should allow domains from several groups', function() {
      acl_config.groups = {
        group1: ['secret2.example.com'],
        group2: ['secret.example.com', 'secret1.example.com']
      };

      var allowed_domains0 = acl_builder.get_allowed_domains('user', ['group1', 'group2']); 
      assert.deepEqual(allowed_domains0, [
        'secret2.example.com',
        'secret.example.com',
        'secret1.example.com',
      ]);
    });

    it('should allow domains from several groups and default policy', function() {
      acl_config.default = ['home.example.com'];
      acl_config.groups = {
        group1: ['secret2.example.com'],
        group2: ['secret.example.com', 'secret1.example.com']
      };

      var allowed_domains0 = acl_builder.get_allowed_domains('user', ['group1', 'group2']); 
      assert.deepEqual(allowed_domains0, [
        'home.example.com',
        'secret2.example.com',
        'secret.example.com',
        'secret1.example.com',
      ]);
    });
  });

  describe('building user group access control matcher', function() {
    it('should allow access to any subdomain', function() {
      var allowed_domains = acl_builder.get_any_domain(); 
      assert(acl_matcher.is_domain_allowed('example.com', allowed_domains));
      assert(acl_matcher.is_domain_allowed('mail.example.com', allowed_domains));
      assert(acl_matcher.is_domain_allowed('test.example.com', allowed_domains));
      assert(acl_matcher.is_domain_allowed('user.mail.example.com', allowed_domains));
      assert(acl_matcher.is_domain_allowed('public.example.com', allowed_domains));
      assert(acl_matcher.is_domain_allowed('example2.com', allowed_domains));
    });
  });

  describe('check access control matching', function() {
    beforeEach(function() {
      acl_config.default = ['home.example.com', '*.public.example.com'];
      acl_config.users = {
        user1: ['user1.example.com', 'user1.mail.example.com']
      };
      acl_config.groups = {
        group1: ['secret2.example.com'],
        group2: ['secret.example.com', 'secret1.example.com']
      };
    });

    it('should allow access to secret.example.com', function() {
      var allowed_domains = acl_builder.get_allowed_domains('user', ['group1', 'group2']); 
      assert(acl_matcher.is_domain_allowed('secret.example.com', allowed_domains));
    });

    it('should deny access to secret3.example.com', function() {
      var allowed_domains = acl_builder.get_allowed_domains('user', ['group1', 'group2']); 
      assert(!acl_matcher.is_domain_allowed('secret3.example.com', allowed_domains));
    });

    it('should allow access to home.example.com', function() {
      var allowed_domains = acl_builder.get_allowed_domains('user', ['group1', 'group2']); 
      assert(acl_matcher.is_domain_allowed('home.example.com', allowed_domains));
    });

    it('should allow access to user1.example.com', function() {
      var allowed_domains = acl_builder.get_allowed_domains('user1', ['group1', 'group2']); 
      assert(acl_matcher.is_domain_allowed('user1.example.com', allowed_domains));
    });

    it('should allow access *.public.example.com', function() {
      var allowed_domains = acl_builder.get_allowed_domains('nouser', []); 
      assert(acl_matcher.is_domain_allowed('user.public.example.com', allowed_domains));
      assert(acl_matcher.is_domain_allowed('test.public.example.com', allowed_domains));
    });
  });
});
