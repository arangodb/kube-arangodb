import React, { Component } from 'react';
import DeploymentOperator from './deployment/DeploymentOperator.js';
import NoOperator from './NoOperator.js';
import Loading from './util/Loading.js';
import api from './api/api.js';
import { Container, Segment, Message } from 'semantic-ui-react';
import './App.css';

const PodInfoView = ({pod, namespace}) => (
  <Segment basic>
    <Message>
      <Message.Header>Kube-ArangoDB</Message.Header>
      <p>Running in Pod <b>{pod}</b> in namespace <b>{namespace}</b>.</p>
    </Message>
  </Segment>
);

const OperatorsView = ({deployment, pod, namespace}) => (
  <div>
    {deployment ? <DeploymentOperator pod-info={<PodInfoView pod={pod} namespace={namespace}/>}/> : <NoOperator />}
  </div>
);

const LoadingView = () => (
  <Container>
    <Loading/>
  </Container>
);

class App extends Component {
  state = {};

  componentDidMount() {
    this.intervalId = setInterval(this.reloadOperators, 5000);
    this.reloadOperators();
  }

  componentWillUnmount() {
    clearInterval(this.intervalId);
  }

  reloadOperators = async() => {
    const operators = await api.get('/api/operators');
    this.setState({operators});
  }

  render() {
    if (this.state.operators) {
      return <OperatorsView 
        deployment={this.state.operators.deployment} 
        pod={this.state.operators.pod} 
        namespace={this.state.operators.namespace} 
      />;
    }
    return (<LoadingView/>);
  }
}

export default App;
