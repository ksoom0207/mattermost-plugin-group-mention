import React from 'react';

interface GroupPillProps {
    groupName: string;
    memberCount?: number;
    visibility?: string;
}

const GroupPill: React.FC<GroupPillProps> = ({groupName, memberCount, visibility}) => {
    const style: React.CSSProperties = {
        display: 'inline-block',
        padding: '2px 6px',
        margin: '0 2px',
        borderRadius: '10px',
        backgroundColor: '#166de0',
        color: '#ffffff',
        fontSize: '0.9em',
        fontWeight: 'bold',
        cursor: 'pointer',
    };

    const tooltipText = memberCount ? `${memberCount} members` : '';
    const visibilityIcon = visibility === 'private' ? '🔒 ' : '';

    return (
        <span
            style={style}
            title={`${visibilityIcon}Group: ${groupName}${tooltipText ? ' - ' + tooltipText : ''}`}
            className='group-mention-pill'
        >
            @{groupName}
        </span>
    );
};

export default GroupPill;
