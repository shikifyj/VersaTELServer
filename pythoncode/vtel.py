import http.client


conn = http.client


# 通过 kubesphere 提供的 API 获取到 VersaTEL 端口
def get_vtelmgt_URL():
    localhost = "10.203.1.158"
    conn = http.client.HTTPConnection(localhost,port=9090)
    conn.request("GET", "/kapis/resources.kubesphere.io/v1alpha3/namespaces/default/services/vtelmgt")
    response = conn.getresponse()
    if not response.closed:
        data = response.read()
        try:
            data = eval(data)
        except Exception:
            data = ""

        nodeport = (data["spec"]["ports"][0]["nodePort"])
        conn.close()
        return f"{localhost}:{nodeport}"


