import React, { Component } from 'react';
//import logo from './logo.svg';
//import './App.css';

class DeploymentRow extends Component {
  render() {
    return (
      <tr>
        <td>{this.props.info.name}</td>
        <td>{this.props.info.mode}</td>
      </tr>
    );
  }
}

export default DeploymentRow;
