import { Header, Menu, Message, Segment } from 'semantic-ui-react';
import React, { Component } from 'react';

import { StyledMenu, StyledContentBox } from '../style/style';
import LogoutContext from '../auth/LogoutContext';
import StorageList from './StorageList';

const ListView = () => (
  <div>
    <Header dividing>
      ArangoLocalStorage resources
    </Header>
    <StorageList/>
  </div>
);

class StorageOperator extends Component {
  render() {
    return (
    <div>
        <LogoutContext.Consumer>
        {doLogout => 
          <StyledMenu fixed="left" vertical>
            <Menu.Item>
              <Menu.Header>Deployment Operator</Menu.Header>
              <Menu.Menu>
                <Menu.Item>
                    Local storages
                </Menu.Item>
                <Menu.Item position="right" onClick={() => doLogout()}>
                    Logout
                </Menu.Item>
              </Menu.Menu>
              {this.props.commonMenuItems}
            </Menu.Item>
          </StyledMenu>
        }
        </LogoutContext.Consumer>
        <StyledContentBox>
        <Segment basic clearing>
            <ListView/>
        </Segment>
        {this.props.podInfoView}
        {(this.props.error) ? <Segment basic><Message error content={this.props.error}/></Segment> : null}
        </StyledContentBox>
    </div>
    );
  }
}

export default StorageOperator;
