# kubectl template
def _kubectl_config(repository_ctx, args):
    kubectl = repository_ctx.path("kubectl")
    kubeconfig_yaml = repository_ctx.path("kubeconfig")
    exec_result = repository_ctx.execute(
        [kubectl, "--kubeconfig", kubeconfig_yaml, "config"] + args,
        environment = {
            # prevent kubectl config to stumble on shared .kube/config.lock file
            "HOME": str(repository_ctx.path(".")),
        },
        quiet = True,
    )
    if exec_result.return_code != 0:
        fail("Error executing kubectl config %s" % " ".join(args))

def _kubeconfig_impl(repository_ctx):
    """Find local kubernetes certificates"""

    # find and symlink kubectl
    kubectl = repository_ctx.which("kubectl")
    if not kubectl:
        fail("Unable to find kubectl executable. PATH=%s" % repository_ctx.path)
    repository_ctx.symlink(kubectl, "kubectl")
    repository_ctx.file(repository_ctx.path("cluster"), content = repository_ctx.attr.cluster, executable = False)

    # TODO: figure out how to use BUILD_USER
    if "USER" in repository_ctx.os.environ:
        user = repository_ctx.os.environ["USER"]
    else:
        exec_result = repository_ctx.execute(["whoami"])
        if exec_result.return_code != 0:
            fail("Error detecting current user")
        user = exec_result.stdout.rstrip()
    token = None
    ca_crt = None
    kubecert_cert = None
    kubecert_key = None
    server = repository_ctx.attr.server

    # check service account first
    serviceaccount = repository_ctx.path("/var/run/secrets/kubernetes.io/serviceaccount")
    if serviceaccount.exists:
        ca_crt = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
        token_file = serviceaccount.get_child("token")
        if token_file.exists:
            exec_result = repository_ctx.execute(["cat", token_file.realpath])
            if exec_result.return_code != 0:
                fail("Error reading user token")
            token = exec_result.stdout.rstrip()

        # use master url from the environemnt
        if "KUBERNETES_SERVICE_HOST" in repository_ctx.os.environ:
            server = "https://%s:%s" % (
                repository_ctx.os.environ["KUBERNETES_SERVICE_HOST"],
                repository_ctx.os.environ["KUBERNETES_SERVICE_PORT"],
            )
        else:
            # fall back to the default
            server = "https://kubernetes.default"
    elif repository_ctx.attr.use_host_config:
        home = repository_ctx.path(repository_ctx.os.environ["HOME"])
        kubeconfig = home.get_child(".kube").get_child("config")
        if repository_ctx.path(kubeconfig).exists:
            repository_ctx.symlink(kubeconfig, repository_ctx.path("kubeconfig"))
        else:
            _kubectl_config(repository_ctx, [
                "set-cluster",
                repository_ctx.attr.cluster,
                "--server",
                server,
            ])
    else:
        home = repository_ctx.path(repository_ctx.os.environ["HOME"])
        certs = home.get_child(".kube").get_child("certs")
        ca_crt = certs.get_child("ca.crt").realpath
        kubecert_cert = certs.get_child("kubecert.cert")
        kubecert_key = certs.get_child("kubecert.key")

    # config set-cluster {cluster} \
    #     --certificate-authority=... \
    #     --server=https://dev3.k8s.tubemogul.info:443 \
    #     --embed-certs",
    if ca_crt:
        _kubectl_config(repository_ctx, [
            "set-cluster",
            repository_ctx.attr.cluster,
            "--server",
            server,
            "--certificate-authority",
            ca_crt,
        ])

    # config set-credentials {user} --token=...",
    if token:
        _kubectl_config(repository_ctx, [
            "set-credentials",
            user,
            "--token",
            token,
        ])

    # config set-credentials {user} --client-certificate=... --embed-certs",
    if kubecert_cert and kubecert_cert.exists:
        _kubectl_config(repository_ctx, [
            "set-credentials",
            user,
            "--client-certificate",
            kubecert_cert.realpath,
        ])

    # config set-credentials {user} --client-key=... --embed-certs",
    if kubecert_key and kubecert_key.exists:
        _kubectl_config(repository_ctx, [
            "set-credentials",
            user,
            "--client-key",
            kubecert_key.realpath,
        ])

    # export repostory contents
    repository_ctx.file("BUILD", """exports_files(["kubeconfig", "kubectl", "cluster"])""", False)

    return {
        "name": repository_ctx.attr.name,
        "cluster": repository_ctx.attr.cluster,
        "server": repository_ctx.attr.server,
        "use_host_config": repository_ctx.attr.use_host_config,
    }

kubeconfig = repository_rule(
    attrs = {
        "cluster": attr.string(),
        "server": attr.string(),
        "use_host_config": attr.bool(),
    },
    environ = [
        "HOME",
        "USER",
        "KUBERNETES_SERVICE_HOST",
        "KUBERNETES_SERVICE_PORT",
    ],
    local = True,
    implementation = _kubeconfig_impl,
)
