from execute import LinstorAPI
import json


class Scheduler():
    def __init__(self):
        pass


    def get_linstor_data(self):
        linstor_api = LinstorAPI()
        # node_data = gevent.spawn(linstor_api.get_node)
        # res_data = gevent.spawn(linstor_api.get_resource)
        # sp_data = gevent.spawn(linstor_api.get_storagepool)
        # gevent.joinall([node_data,res_data,sp_data])

        node_data = linstor_api.get_node()
        res_data = linstor_api.get_resource()
        sp_data = linstor_api.get_storagepool()

        print("调用了linstor api")

        return {'node_data':node_data,'res_data':res_data,'sp_data':sp_data}



class VersaTEL():
    def process_data_node(self):
        data = Scheduler().get_linstor_data()
        data2 = {
            "code": 0,
            "count": 1,
            "data": [
                {
                    "addr": "10.203.1.155:3366 (PLAIN)",
                    "name": "vince1",
                    "node_type": "COMBINED",
                    "res_num": "1",
                    "status": "ONLINE",
                    "stp_num": "2"
                }
            ]
        }

        result_str = json.dumps(data2)
        return result_str
