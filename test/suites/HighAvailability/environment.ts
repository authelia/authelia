const composeFiles = [
  'docker-compose.yml',
  'example/compose/authelia/docker-compose.backend.yml',
  'example/compose/authelia/docker-compose.frontend.yml',
  'example/compose/mariadb/docker-compose.yml',
  'example/compose/redis/docker-compose.yml',
  'example/compose/nginx/backend/docker-compose.yml',
  'example/compose/nginx/portal/docker-compose.yml',
  'example/compose/smtp/docker-compose.yml',
  'example/compose/httpbin/docker-compose.yml',
  'example/compose/ldap/docker-compose.admin.yml', // This is just used for administration, not for testing.
  'example/compose/ldap/docker-compose.yml'
]

export {
  composeFiles,
};