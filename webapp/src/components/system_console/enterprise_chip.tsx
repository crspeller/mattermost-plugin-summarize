import React from 'react';
import styled from 'styled-components';

//eslint-disable-next-line import/no-unresolved -- react-bootstrap is external
import {OverlayTrigger, Tooltip} from 'react-bootstrap';

const Chip = styled.div`
    position: relative;
    display: flex;
    align-items: center;
    padding: 3px 8px 3px 22px;
    margin-left: 8px;
    background: var(--button-bg);
    border-radius: 10px;
    height: 20px;

    font-size: 10px;
    font-weight: 600;
    line-height: 15px;
    color: var(--center-channel-bg);

    &:before {
        left: 7px;
        top: 3px;
        position: absolute;
        content: '\f030b';
        font-size: 12px;
        font-family: 'compass-icons', mattermosticons;
        -webkit-font-smoothing: antialiased;
        -moz-osx-font-smoothing: grayscale;
    }
`;

const MainText = styled.div`
	font-size: 12px;
	fong-weight: 600;
	line-height: 15px;
`;

const SubText = styled.div`
	font-size: 11px;
	font-weight: 600;
	line-height: 16px;
	letter-spacing: 0.22px;
	opacity: 0.56;
`;

type Props = {
    subtext?: string;
};

const EnterpriseChip = (props: Props) => {
    return (
        <OverlayTrigger
            placement='top'
            overlay={
                <Tooltip>
                    <MainText>{'Enterprise feature'}</MainText>
                    <SubText>{props.subtext}</SubText>
                </Tooltip>
            }
        >
            <Chip>
                {'Enterprise'}
            </Chip>
        </OverlayTrigger>
    );
};

export default EnterpriseChip;
