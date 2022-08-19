import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';

const client = new grpc.Client()
client.load([], 'api/proto/broker.proto');

export default () => {
    client.connect('localhost:8080', {
        'plaintext': true,
        'timeout': '10s',
    });

    const data1 = {
        "body": "testMessage",
        "expirationSeconds": 10,
        "subject": "test"
    };
    // const data2 = {
    //     "body": "testMessage",
    //     "expirationSeconds": 10,
    //     "subject": "test2"
    // };

    const response1 = client.invoke('broker.Broker/Publish', data1);
    // const response2 = client.invoke('broker.Broker/Publish', data2);

    check(response1, {
        'status is OK': (r) => r && r.status === grpc.StatusOK,
    });
    // check(response2, {
    //     'status is OK': (r) => r && r.status === grpc.StatusOK,
    // });
    
    // console.log(JSON.stringify(response1.message));
    // console.log(JSON.stringify(response2.message));

    // client.close();
}