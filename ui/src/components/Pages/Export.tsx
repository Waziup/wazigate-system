import React, { Component } from "react";

import {
    MDBBtn,
    MDBIcon,
    MDBCardText,
    MDBRow,
    MDBContainer
  } from "mdbreact";

class PageExport extends React.Component {
    render() {
        return (
            <>                 
            <MDBContainer>
              <MDBRow>
                <MDBCardText>You can export the data of all sensors and actuators to one CSV file:</MDBCardText>
                <MDBBtn disabled={false} href="exportall" target="_blank">
                      <MDBIcon
                        icon="all_match"
                        className="ml-2"
                        size="1x"
                      />{" "}
                      Export data to one CSV file
                </MDBBtn>
              </MDBRow>
              <MDBRow>
                <MDBCardText>You can export the data of all sensors and actuators to a tree of CSV files:</MDBCardText>
                <MDBBtn disabled={false} href="exporttree" target="_blank">
                      <MDBIcon
                        icon="account_tree"
                        className="ml-2"
                        size="1x"
                      />{" "}
                      Export data to tree of CSV files
                </MDBBtn>
              </MDBRow>
            </MDBContainer>
            </>
        );
    }   
}

export default PageExport