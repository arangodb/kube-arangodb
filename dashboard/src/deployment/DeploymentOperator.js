import React, { Component } from 'react';
import DeploymentList from './DeploymentList.js';
import { Header, Segment } from 'semantic-ui-react';
//import logo from './logo.svg';
//import './App.css';

class DeploymentOperator extends Component {
  render() {
    return (
      <Segment basic>
        <Header dividing>
          ArangoDeployments....
        </Header>
        <DeploymentList/>
      </Segment>
    );
  }
}

export default DeploymentOperator;
