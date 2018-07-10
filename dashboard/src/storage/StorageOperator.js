import { Header, Menu, Message, Segment } from 'semantic-ui-react';
import React, { Component } from 'react';
import styled from 'react-emotion';

import LogoutContext from '../auth/LogoutContext';
import StorageList from './StorageList';

const StyledMenu = styled(Menu)`
  width: 15rem !important;
  @media (max-width: 768px) {
    width: 10rem !important;
  }
`;

const StyledContentBox = styled('div')`
  margin-left: 15rem;
  @media (max-width: 768px) {
    margin-left: 10rem;
  }
`;

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
                Local storages
            </Menu.Item>
            <Menu.Item position="right" onClick={() => doLogout()}>
                Logout
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
