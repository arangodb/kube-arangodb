import React, { Component } from 'react';
import { apiGet } from '../api/api.js';
import DeploymentRow from './DeploymentRow.js';
//import logo from './logo.svg';
//import './App.css';

class DeploymentList extends Component {
  constructor() {
    super();
    this.state = {items:[]};
  }

  async componentDidMount() {
    this.reloadDeployments();
  }

  async reloadDeployments() {
    const result = await apiGet('/api/deployment');
    this.setState({items:result.deployments});
  }

  render() {
    setTimeout(this.reloadDeployments.bind(this), 5000);
    const items = this.state.items;
    return (
      <table>
        <tbody>
          {
            (items) ? items.map((item) => <DeploymentRow key={item.name} info={item}/>) : <p>No items</p>
          }
        </tbody>
      </table>
    );
  }
}

export default DeploymentList;
