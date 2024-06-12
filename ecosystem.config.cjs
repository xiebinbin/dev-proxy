module.exports = {
  apps: [
    {
      name: 'mirror-proxy',
      script: './server',
      args: ' -h 127.0.0.1 -p 8080',
      env: {
        GIN_MODE: 'release'
      }
    }
  ]
}
