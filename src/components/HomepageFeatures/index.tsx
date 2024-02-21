import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  Svg: React.ComponentType<React.ComponentProps<'svg'>>;
  description: JSX.Element;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'Simple and Readable Syntax',
    Svg: require('@site/static/img/gocodesvg.svg').default,
    description: (
      <>
        Go's clean and concise syntax reduces the cognitive load, making it easy for developers to understand and maintain codebases.
      </>
    ),
  },
  {
    title: 'Fast Compilation',
    Svg: require('@site/static/img/gofastsvg.svg').default,
    description: (
      <>
        With its fast compilation times, Go enables rapid development cycles, making it ideal for building scalable and maintainable applications.&apos;
        <code></code>
      </>
    ),
  },
  {
    title: 'Concurrency Built In',
    Svg: require('@site/static/img/undraw_docusaurus_react.svg').default,
    description: (
      <>
        Go makes it easy to write concurrent programs through goroutines and channels, allowing for efficient use of multiple CPU cores.
      </>
    ),
  },
];

function Feature({title, Svg, description}: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <Svg className={styles.featureSvg} role="img" />
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): JSX.Element {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
